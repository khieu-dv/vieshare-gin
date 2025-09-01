package routers

import (
	"net/http"
	"runtime"

	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// SetupStaticRoutes sets up static file serving and documentation routes
func SetupStaticRoutes(r *gin.Engine) {
	// Swagger documentation
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	// Static files
	r.LoadHTMLGlob("./public/html/*")
	r.Static("/public", "./public")

	// Home page
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"ginBoilerplateVersion": "v0.03",
			"goVersion":             runtime.Version(),
		})
	})

	// 404 handler
	r.NoRoute(func(c *gin.Context) {
		c.HTML(404, "404.html", gin.H{})
	})
}