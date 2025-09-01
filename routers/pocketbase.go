package routers

import (
	"github.com/VieShare/vieshare-gin/controllers"
	"github.com/gin-gonic/gin"
)

// SetupPocketBaseRoutes sets up PocketBase-compatible API routes
func SetupPocketBaseRoutes(r *gin.RouterGroup) {
	pb := new(controllers.PocketBaseController)
	
	// Health check
	r.GET("/health", pb.Health)
	
	// Collections CRUD operations
	collections := r.Group("/collections/:collection")
	{
		collections.GET("/records", pb.ListRecords)
		collections.GET("/records/:id", pb.GetRecord)
		collections.POST("/records", pb.CreateRecord)
		collections.PATCH("/records/:id", pb.UpdateRecord)
		collections.DELETE("/records/:id", pb.DeleteRecord)
	}
}