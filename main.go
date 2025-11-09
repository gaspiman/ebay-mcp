package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
)

var (
	// oauthConf is the global OAuth2 configuration for eBay.
	oauthConf *oauth2.Config

	// ebayClientID is your eBay App's Client ID.
	ebayClientID string

	// ebayClientSecret is your eBay App's Client Secret.
	ebayClientSecret string

	// ebayAPIHost is the target eBay API host (e.g., "api.ebay.com").
	ebayAPIHost string

	// stateStore links the 'state' string to OpenAI's 'redirect_uri'.
	// For production, use a proper store (e.g., Redis) with a short TTL.
	stateStore = make(map[string]string)
)

// ### Main Server Setup (with Autocert) ####################################

func main() {
	// 0. Load .env file (if it exists)
	// This will load variables from .env file into the environment.
	// If the file doesn't exist, it will silently continue (good for production).
	if err := godotenv.Load("../.env"); err != nil {
		log.Println("No .env file found, using existing environment variables")
	}

	log.Println("Loaded Env")

	// 1. Load configuration from Environment Variables
	ebayClientID = os.Getenv("EBAY_CLIENT_ID")
	ebayClientSecret = os.Getenv("EBAY_CLIENT_SECRET")
	appRedirectURL := os.Getenv("APP_REDIRECT_URL") // <-- MUST be "https://ebayai.dev/callback"
	ebayScopes := os.Getenv("EBAY_SCOPES")          // Space-separated list of scopes
	ebayAPIHost = os.Getenv("EBAY_API_HOST")        // "api.ebay.com" or "api.sandbox.ebay.com"
	ebayAuthURL := os.Getenv("EBAY_AUTH_URL")       // "https://auth.ebay.com/oauth2/authorize"
	ebayTokenURL := os.Getenv("EBAY_TOKEN_URL")     // "https://api.ebay.com/identity/v1/oauth2/token"
	sslCertFile := os.Getenv("SSL_CERTFILE")        // Path to SSL certificate file
	sslKeyFile := os.Getenv("SSL_KEYFILE")          // Path to SSL key file

	// !! CRITICAL !!
	// Validate the APP_REDIRECT_URL for production
	if appRedirectURL != "https://ebayai.dev/callback" {
		log.Fatalf("FATAL: APP_REDIRECT_URL must be set to 'https://ebayai.dev/callback' for production. Found: %s", appRedirectURL)
	}

	// Basic validation
	if ebayClientID == "" || ebayClientSecret == "" || ebayScopes == "" || ebayAPIHost == "" || ebayAuthURL == "" || ebayTokenURL == "" {
		log.Fatal("Error: Missing required environment variables. \n" +
			"Please set: EBAY_CLIENT_ID, EBAY_CLIENT_SECRET, APP_REDIRECT_URL, EBAY_SCOPES, EBAY_API_HOST, EBAY_AUTH_URL, EBAY_TOKEN_URL")
	}

	// Validate SSL certificate paths
	if sslCertFile == "" || sslKeyFile == "" {
		log.Fatal("Error: Missing SSL certificate configuration. \n" +
			"Please set: SSL_CERTFILE, SSL_KEYFILE")
	}

	// 2. Initialize the oauth2.Config
	// This config is for the flow between YOUR server and EBAY.
	oauthConf = &oauth2.Config{
		ClientID:     ebayClientID,
		ClientSecret: ebayClientSecret,
		RedirectURL:  appRedirectURL, // This is YOUR /callback endpoint
		Scopes:       strings.Split(ebayScopes, " "),
		Endpoint: oauth2.Endpoint{
			AuthURL:  ebayAuthURL,
			TokenURL: ebayTokenURL,
		},
	}

	// 3. Define HTTP handlers
	// We create a router (mux) to hold all our handlers.
	mux := http.NewServeMux()
	mux.HandleFunc("/authorize", handleAuthorize) // OpenAI starts here
	mux.HandleFunc("/callback", handleCallback)   // eBay redirects user here
	mux.HandleFunc("/token", handleToken)         // OpenAI calls this to get token
	mux.HandleFunc("/proxy/", handleProxy)        // OpenAI calls this for API requests
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "eBay GPT Action Proxy is running securely on https://ebayai.dev")
	})

	// 4. Configure the main HTTPS server using existing certificates
	// Wrap the mux with logging middleware to log all requests
	server := &http.Server{
		Addr:    ":443",                 // Listen on port 443
		Handler: loggingMiddleware(mux), // Use the router wrapped with logging
	}

	// 5. Start the main HTTPS server with existing Let's Encrypt certificates
	log.Println("Starting eBay GPT proxy server on https://ebayai.dev (port 443)...")
	log.Printf("Using SSL certificate: %s", sslCertFile)
	log.Printf("Using SSL key: %s", sslKeyFile)
	if err := server.ListenAndServeTLS(sslCertFile, sslKeyFile); err != nil {
		log.Fatalf("HTTPS server error: %v", err)
	}
}

// ### OAuth Handlers (OpenAI Flow) ###########################################

// handleAuthorize: Called by OpenAI to start the login flow.
// It receives OpenAI's redirect_uri and state.
func handleAuthorize(w http.ResponseWriter, r *http.Request) {
	// 1. Get parameters from OpenAI
	openAIRedirectURI := r.URL.Query().Get("redirect_uri")
	state := r.URL.Query().Get("state")

	if openAIRedirectURI == "" || state == "" {
		http.Error(w, "Missing required parameters: redirect_uri and state", http.StatusBadRequest)
		return
	}

	// 2. Store OpenAI's redirect_uri, keyed by state
	log.Printf("Storing state: %s -> %s", state, openAIRedirectURI)
	stateStore[state] = openAIRedirectURI

	// 3. Generate the eBay auth URL and redirect the user's browser
	// We use AccessTypeOffline to request a refresh token
	url := oauthConf.AuthCodeURL(state, oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// handleCallback: Called by eBay after the user grants consent.
// It receives the 'code' and 'state' from eBay.
func handleCallback(w http.ResponseWriter, r *http.Request) {
	// 1. Get code and state from eBay's redirect
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	if code == "" {
		http.Error(w, "Code not found", http.StatusBadRequest)
		return
	}

	// 2. Retrieve the original OpenAI redirect_uri from our store
	openAIRedirectURI, ok := stateStore[state]
	if !ok {
		log.Println("Invalid or expired OAuth state received")
		http.Error(w, "Invalid state", http.StatusBadRequest)
		return
	}
	delete(stateStore, state) // State is single-use

	// 3. Redirect back to OpenAI's callback URL, passing along the code.
	// OpenAI will then call our /token endpoint.
	redirectURL, err := url.Parse(openAIRedirectURI)
	if err != nil {
		log.Printf("Invalid OpenAI redirect_uri: %v", err)
		http.Error(w, "Invalid redirect_uri", http.StatusInternalServerError)
		return
	}

	q := redirectURL.Query()
	q.Set("code", code)
	q.Set("state", state)
	redirectURL.RawQuery = q.Encode()

	log.Printf("Redirecting back to OpenAI: %s", redirectURL.String())
	http.Redirect(w, r, redirectURL.String(), http.StatusTemporaryRedirect)
}

// handleToken: Called by OpenAI's backend to exchange the code for a token
// or to refresh an existing token.
// This endpoint is *not* called by a user's browser.
func handleToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse the form data from OpenAI
	if err := r.ParseForm(); err != nil {
		log.Printf("Failed to parse form: %v", err)
		http.Error(w, "Failed to parse request body", http.StatusBadRequest)
		return
	}

	// Extract parameters from OpenAI's request
	code := r.Form.Get("code")
	grantType := r.Form.Get("grant_type")
	refreshToken := r.Form.Get("refresh_token")
	redirectURI := r.Form.Get("redirect_uri")

	log.Printf("Token request - grant_type: %s, has_code: %v, has_refresh_token: %v, redirect_uri: %s",
		grantType, code != "", refreshToken != "", redirectURI)

	// Build the form data to send to eBay with correct parameters
	formData := url.Values{}

	if grantType == "refresh_token" && refreshToken != "" {
		// Handle refresh token flow
		// eBay requires the redirect_uri and scope even for refresh tokens
		formData.Set("grant_type", "refresh_token")
		formData.Set("refresh_token", refreshToken)
		formData.Set("redirect_uri", oauthConf.RedirectURL)
		// Include the same scopes that were used in the original authorization
		formData.Set("scope", strings.Join(oauthConf.Scopes, " "))
	} else if code != "" {
		// Handle authorization code flow
		formData.Set("grant_type", "authorization_code")
		formData.Set("code", code)
		// IMPORTANT: Must use OUR redirect_uri (not OpenAI's) because that's what
		// we used in the authorization request and what's registered with eBay
		formData.Set("redirect_uri", oauthConf.RedirectURL)
	} else {
		log.Printf("Invalid token request: missing code or refresh_token")
		http.Error(w, "Missing required parameters", http.StatusBadRequest)
		return
	}

	// Log what we're sending to eBay
	log.Printf("Sending to eBay token endpoint: %s", formData.Encode())

	// Create a new request to eBay's token endpoint
	proxyReq, err := http.NewRequestWithContext(context.Background(), "POST",
		oauthConf.Endpoint.TokenURL, strings.NewReader(formData.Encode()))
	if err != nil {
		log.Printf("Failed to create proxy request: %v", err)
		http.Error(w, "Failed to create proxy request", http.StatusInternalServerError)
		return
	}

	// --- This is the critical part ---
	// Add the Basic Auth header using the server's *secret* credentials
	auth := base64.StdEncoding.EncodeToString([]byte(ebayClientID + ":" + ebayClientSecret))
	proxyReq.Header.Set("Authorization", "Basic "+auth)

	// Set the Content-Type header
	proxyReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Send the request to eBay
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(proxyReq)
	if err != nil {
		log.Printf("Failed to send request to eBay token endpoint: %v", err)
		http.Error(w, "Failed to send request to token endpoint", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Log the response status from eBay
	log.Printf("eBay token endpoint response: %d", resp.StatusCode)

	// Read the response body from eBay
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read eBay response: %v", err)
		http.Error(w, "Failed to read token response", http.StatusInternalServerError)
		return
	}

	// If there was an error, log and return it
	if resp.StatusCode >= 400 {
		log.Printf("eBay error response: %s", string(bodyBytes))
		copyHeaders(w.Header(), resp.Header)
		w.WriteHeader(resp.StatusCode)
		w.Write(bodyBytes)
		return
	}

	// Parse the successful token response to modify token_type
	var tokenResponse map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &tokenResponse); err != nil {
		log.Printf("Failed to parse eBay token response: %v", err)
		// If we can't parse it, just return as-is
		copyHeaders(w.Header(), resp.Header)
		w.WriteHeader(resp.StatusCode)
		w.Write(bodyBytes)
		return
	}

	// eBay returns "token_type": "User Access Token" but OAuth 2.0 standard expects "Bearer"
	// Normalize the token_type to "Bearer" for compatibility with ChatGPT
	if _, ok := tokenResponse["token_type"]; ok {
		log.Printf("Original token_type from eBay: %v", tokenResponse["token_type"])
		tokenResponse["token_type"] = "Bearer"
	}

	// Re-encode the modified response
	modifiedBody, err := json.Marshal(tokenResponse)
	if err != nil {
		log.Printf("Failed to encode modified token response: %v", err)
		// If we can't encode it, return original
		copyHeaders(w.Header(), resp.Header)
		w.WriteHeader(resp.StatusCode)
		w.Write(bodyBytes)
		return
	}

	log.Printf("Modified token response: %s", string(modifiedBody))

	// Send the modified response to OpenAI
	copyHeaders(w.Header(), resp.Header)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(modifiedBody)))
	w.WriteHeader(resp.StatusCode)
	w.Write(modifiedBody)
}

// ### API Proxy Handler (OpenAI Flow) ########################################

// handleProxy: Called by OpenAI for all API requests.
// It expects OpenAI to provide a valid 'Authorization: Bearer <token>' header.
func handleProxy(w http.ResponseWriter, r *http.Request) {
	// 1. Get the token from the Authorization header sent by OpenAI
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header required", http.StatusUnauthorized)
		return
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		http.Error(w, "Invalid Authorization header: must be 'Bearer {token}'", http.StatusUnauthorized)
		return
	}
	accessToken := parts[1]

	// 2. Create the reverse proxy to eBay
	targetURL, _ := url.Parse("https://" + ebayAPIHost)
	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	// Enable HTTP/2 properly for eBay API
	// eBay requires HTTP/2, so we need to enable it with proper configuration
	proxy.Transport = &http.Transport{
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		ResponseHeaderTimeout: 45 * time.Second, // Increased timeout for eBay API
		IdleConnTimeout:       90 * time.Second, // Keep idle connections for 90 seconds
		MaxIdleConns:          100,              // Maximum idle connections
		MaxIdleConnsPerHost:   10,               // Maximum idle connections per host
		MaxConnsPerHost:       50,               // Maximum total connections per host
		DisableKeepAlives:     false,            // Enable keep-alives for better performance
		ForceAttemptHTTP2:     true,             // Enable HTTP/2
	}

	// Store the path we'll actually send to eBay for logging
	strippedPath := strings.TrimPrefix(r.URL.Path, "/proxy")

	// 3. Set the Director to modify the request *before* it's sent to eBay
	proxy.Director = func(req *http.Request) {
		// Set the target host and scheme
		req.URL.Scheme = targetURL.Scheme
		req.URL.Host = targetURL.Host
		req.Host = targetURL.Host // Set the host header

		// Set the correct API path by stripping our /proxy prefix
		// e.g., /proxy/sell/inventory/v1/item_summary/search -> /sell/inventory/v1/item_summary/search
		req.URL.Path = strippedPath
		req.URL.RawQuery = r.URL.RawQuery // Preserve query parameters

		log.Printf("Proxying to eBay: %s %s%s?%s", req.Method, req.URL.Host, req.URL.Path, req.URL.RawQuery)

		// --- This is the critical part ---
		// Add the OAuth Authorization header using the token OpenAI sent
		req.Header.Set("Authorization", "Bearer "+accessToken)

		// Set required headers for eBay API
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Content-Type", "application/json")

		// Clean up headers not meant for eBay
		// Remove all OpenAI/ChatGPT specific headers that might confuse eBay
		req.Header.Del("Cookie")
		req.Header.Del("Openai-Conversation-Id")
		req.Header.Del("Openai-Ephemeral-User-Id")
		req.Header.Del("Openai-Gpt-Id")
		req.Header.Del("Traceparent")
		req.Header.Del("Tracestate")
		req.Header.Del("X-Datadog-Parent-Id")
		req.Header.Del("X-Datadog-Sampling-Priority")
		req.Header.Del("X-Datadog-Tags")
		req.Header.Del("X-Datadog-Trace-Id")
		req.Header.Del("X-Request-Id")

		// Set a clean User-Agent
		req.Header.Set("User-Agent", "eBay-Proxy/1.0")

		// Log the outgoing headers (mask the token for security)
		maskedHeaders := make(map[string][]string)
		for k, v := range req.Header {
			if k == "Authorization" {
				maskedHeaders[k] = []string{"Bearer ***MASKED***"}
			} else {
				maskedHeaders[k] = v
			}
		}
		log.Printf("Request headers to eBay: %v", maskedHeaders)
	}

	// 4. Add response modifier to log responses from eBay
	proxy.ModifyResponse = func(resp *http.Response) error {
		log.Printf("Received response from eBay: Status %d %s", resp.StatusCode, resp.Status)
		log.Printf("Response headers from eBay: %v", resp.Header)

		// If there's an error status, log the response body
		if resp.StatusCode >= 400 {
			bodyBytes, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Printf("Failed to read error response body: %v", err)
				return err
			}
			log.Printf("eBay API error response body: %s", string(bodyBytes))

			// Restore the body for the client
			resp.Body = io.NopCloser(strings.NewReader(string(bodyBytes)))
		}

		return nil
	}

	// 5. Add error handler to log proxy errors
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("PROXY ERROR: %v", err)
		log.Printf("Failed request: %s %s", r.Method, r.URL.String())
		log.Printf("Target was: %s%s", targetURL.Host, strippedPath)
		http.Error(w, fmt.Sprintf("Proxy error: %v", err), http.StatusBadGateway)
	}

	// 6. Serve the request with timing
	log.Printf("Proxying %s request to %s%s", r.Method, targetURL.Host, strippedPath)
	startTime := time.Now()
	proxy.ServeHTTP(w, r)
	elapsed := time.Since(startTime)
	log.Printf("eBay API request completed in %v", elapsed)
}

// ### Helper Functions #######################################################

// loggingMiddleware logs all incoming HTTP requests
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Log request details
		log.Printf("[REQUEST] %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
		log.Printf("[REQUEST] Headers: %v", r.Header)
		log.Printf("[REQUEST] Query: %v", r.URL.RawQuery)

		// Call the next handler
		next.ServeHTTP(w, r)

		// Log request completion time
		duration := time.Since(start)
		log.Printf("[REQUEST] Completed %s %s in %v", r.Method, r.URL.Path, duration)
	})
}

// copyHeaders copies all headers from src to dst.
func copyHeaders(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}
