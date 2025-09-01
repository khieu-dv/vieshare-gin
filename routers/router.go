package routers

import (
	"github.com/VieShare/vieshare-gin/controllers"
	"github.com/gin-gonic/gin"
)

// SetupV1Routes sets up all v1 API routes
func SetupV1Routes(r *gin.RouterGroup, auth gin.HandlerFunc) {
	// User routes
	setupUserRoutes(r)
	
	// Auth routes  
	setupAuthRoutes(r)
	
	// Article routes (with auth middleware)
	setupArticleRoutes(r, auth)
}

// setupUserRoutes sets up user-related routes
func setupUserRoutes(r *gin.RouterGroup) {
	user := new(controllers.UserController)
	
	r.POST("/user/login", user.Login)
	r.POST("/user/register", user.Register)
	r.GET("/user/logout", user.Logout)
}

// setupAuthRoutes sets up authentication routes
func setupAuthRoutes(r *gin.RouterGroup) {
	auth := new(controllers.AuthController)
	
	// Refresh the token when needed to generate new access_token and refresh_token for the user
	r.POST("/token/refresh", auth.Refresh)
}

// setupArticleRoutes sets up article-related routes with authentication
// Note: Article functionality has been replaced with PocketBase collections
func setupArticleRoutes(r *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	// Articles are now handled via PocketBase collections API
	// Example: /api/collections/articles/records
	// This function is kept for backward compatibility but routes are disabled
}