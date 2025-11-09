package controllers

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"ebay-mcp/backend/config"
	"ebay-mcp/backend/database"
	"ebay-mcp/backend/models"
	"ebay-mcp/backend/utils"

	"github.com/gin-gonic/gin"
)

type OAuthController struct {
	config *config.Config
}

func NewOAuthController(cfg *config.Config) *OAuthController {
	return &OAuthController{config: cfg}
}

// Authorize handles the OAuth authorization endpoint
// GET /oauth/authorize?client_id=xxx&redirect_uri=xxx&response_type=code&scope=xxx&state=xxx
func (ctrl *OAuthController) Authorize(c *gin.Context) {
	clientID := c.Query("client_id")
	redirectURI := c.Query("redirect_uri")
	responseType := c.Query("response_type")
	scope := c.Query("scope")
	state := c.Query("state")

	// Validate required parameters
	if clientID == "" || redirectURI == "" || responseType != "code" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request parameters"})
		return
	}

	// Verify client exists
	var client models.OAuthClient
	if err := database.DB.Where("id = ?", clientID).First(&client).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid client_id"})
		return
	}

	// Verify redirect_uri is registered for this client
	var redirectURIs []string
	if err := json.Unmarshal([]byte(client.RedirectURIs), &redirectURIs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid client configuration"})
		return
	}

	validRedirectURI := false
	for _, uri := range redirectURIs {
		if uri == redirectURI {
			validRedirectURI = true
			break
		}
	}

	if !validRedirectURI {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid redirect_uri"})
		return
	}

	// Check if user is authenticated
	userID, exists := c.Get("user_id")
	if !exists {
		// Redirect to login with return URL
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":      "authentication_required",
			"login_url":  ctrl.config.FrontendURL + "/login",
			"client_id":  clientID,
			"client_name": client.Name,
		})
		return
	}

	// Return consent screen data
	c.JSON(http.StatusOK, gin.H{
		"client_id":    clientID,
		"client_name":  client.Name,
		"redirect_uri": redirectURI,
		"scope":        scope,
		"state":        state,
		"user_id":      userID,
	})
}

// AuthorizeConsent handles the user's consent decision
// POST /oauth/authorize/consent
func (ctrl *OAuthController) AuthorizeConsent(c *gin.Context) {
	var req struct {
		ClientID    string `json:"client_id" binding:"required"`
		RedirectURI string `json:"redirect_uri" binding:"required"`
		Scope       string `json:"scope"`
		State       string `json:"state"`
		Approved    bool   `json:"approved"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get authenticated user
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Check if user denied
	if !req.Approved {
		c.JSON(http.StatusOK, gin.H{
			"redirect_url": req.RedirectURI + "?error=access_denied&state=" + req.State,
		})
		return
	}

	// Generate authorization code
	code, err := utils.GenerateRandomToken(32)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate code"})
		return
	}

	// Save authorization code to database
	authCode := models.OAuthAuthorizationCode{
		Code:        code,
		ClientID:    req.ClientID,
		UserID:      userID.(uint),
		RedirectURI: req.RedirectURI,
		Scope:       req.Scope,
		ExpiresAt:   time.Now().Add(10 * time.Minute), // Code valid for 10 minutes
		Used:        false,
	}

	if err := database.DB.Create(&authCode).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create authorization code"})
		return
	}

	// Build redirect URL with code
	redirectURL := req.RedirectURI + "?code=" + code
	if req.State != "" {
		redirectURL += "&state=" + req.State
	}

	c.JSON(http.StatusOK, gin.H{
		"redirect_url": redirectURL,
	})
}

// Token handles the OAuth token endpoint
// POST /oauth/token
func (ctrl *OAuthController) Token(c *gin.Context) {
	var req struct {
		GrantType    string `form:"grant_type" binding:"required"`
		Code         string `form:"code"`
		RedirectURI  string `form:"redirect_uri"`
		ClientID     string `form:"client_id" binding:"required"`
		ClientSecret string `form:"client_secret" binding:"required"`
		RefreshToken string `form:"refresh_token"`
	}

	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "error_description": err.Error()})
		return
	}

	// Verify client credentials
	var client models.OAuthClient
	if err := database.DB.Where("id = ? AND client_secret = ?", req.ClientID, req.ClientSecret).First(&client).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_client"})
		return
	}

	switch req.GrantType {
	case "authorization_code":
		ctrl.handleAuthorizationCodeGrant(c, req.Code, req.RedirectURI, req.ClientID)
	case "refresh_token":
		ctrl.handleRefreshTokenGrant(c, req.RefreshToken, req.ClientID)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported_grant_type"})
	}
}

func (ctrl *OAuthController) handleAuthorizationCodeGrant(c *gin.Context, code, redirectURI, clientID string) {
	// Find and validate authorization code
	var authCode models.OAuthAuthorizationCode
	if err := database.DB.Where("code = ? AND client_id = ? AND redirect_uri = ? AND used = ? AND expires_at > ?",
		code, clientID, redirectURI, false, time.Now()).First(&authCode).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_grant"})
		return
	}

	// Mark code as used
	database.DB.Model(&authCode).Update("used", true)

	// Generate access token
	accessToken, err := utils.GenerateRandomToken(32)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server_error"})
		return
	}

	// Generate refresh token
	refreshToken, err := utils.GenerateRandomToken(32)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server_error"})
		return
	}

	// Save tokens to database
	accessTokenModel := models.OAuthAccessToken{
		Token:     accessToken,
		ClientID:  clientID,
		UserID:    authCode.UserID,
		Scope:     authCode.Scope,
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}

	refreshTokenModel := models.OAuthRefreshToken{
		Token:     refreshToken,
		ClientID:  clientID,
		UserID:    authCode.UserID,
		Scope:     authCode.Scope,
		ExpiresAt: time.Now().Add(30 * 24 * time.Hour), // 30 days
	}

	if err := database.DB.Create(&accessTokenModel).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server_error"})
		return
	}

	if err := database.DB.Create(&refreshTokenModel).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server_error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken,
		"token_type":    "Bearer",
		"expires_in":    3600,
		"refresh_token": refreshToken,
		"scope":         authCode.Scope,
	})
}

func (ctrl *OAuthController) handleRefreshTokenGrant(c *gin.Context, refreshToken, clientID string) {
	// Find and validate refresh token
	var refreshTokenModel models.OAuthRefreshToken
	if err := database.DB.Where("token = ? AND client_id = ? AND expires_at > ?",
		refreshToken, clientID, time.Now()).First(&refreshTokenModel).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_grant"})
		return
	}

	// Generate new access token
	accessToken, err := utils.GenerateRandomToken(32)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server_error"})
		return
	}

	// Save new access token
	accessTokenModel := models.OAuthAccessToken{
		Token:     accessToken,
		ClientID:  clientID,
		UserID:    refreshTokenModel.UserID,
		Scope:     refreshTokenModel.Scope,
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}

	if err := database.DB.Create(&accessTokenModel).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server_error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token": accessToken,
		"token_type":   "Bearer",
		"expires_in":   3600,
		"scope":        refreshTokenModel.Scope,
	})
}

// UserInfo returns user information for a valid access token
// GET /oauth/userinfo
func (ctrl *OAuthController) UserInfo(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_request"})
		return
	}

	// Extract token
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_request"})
		return
	}

	token := parts[1]

	// Find and validate access token
	var accessToken models.OAuthAccessToken
	if err := database.DB.Where("token = ? AND expires_at > ?", token, time.Now()).
		Preload("User").First(&accessToken).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"sub":   accessToken.UserID,
		"email": accessToken.User.Email,
		"name":  accessToken.User.Name,
	})
}
