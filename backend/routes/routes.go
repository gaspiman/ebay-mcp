package routes

import (
	"ebay-mcp/backend/config"
	"ebay-mcp/backend/controllers"
	"ebay-mcp/backend/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine, cfg *config.Config) {
	// Initialize controllers
	authController := controllers.NewAuthController(cfg)
	oauthController := controllers.NewOAuthController(cfg)

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Auth routes (public)
	auth := router.Group("/api/auth")
	{
		auth.POST("/register", authController.Register)
		auth.POST("/login", authController.Login)
	}

	// Protected auth routes
	authProtected := router.Group("/api/auth")
	authProtected.Use(middleware.AuthMiddleware(cfg))
	{
		authProtected.GET("/profile", authController.GetProfile)
	}

	// OAuth routes
	oauth := router.Group("/oauth")
	{
		// Authorization endpoint (requires authentication)
		oauthProtected := oauth.Group("")
		oauthProtected.Use(middleware.AuthMiddleware(cfg))
		{
			oauthProtected.GET("/authorize", oauthController.Authorize)
			oauthProtected.POST("/authorize/consent", oauthController.AuthorizeConsent)
		}

		// Token endpoint (public - uses client credentials)
		oauth.POST("/token", oauthController.Token)

		// UserInfo endpoint (requires OAuth access token)
		oauth.GET("/userinfo", oauthController.UserInfo)
	}
}
