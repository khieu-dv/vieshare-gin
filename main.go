package main

import (
	"log"
	"os"

	"github.com/VieShare/vieshare-gin/db"
	_ "github.com/VieShare/vieshare-gin/docs"
	"github.com/VieShare/vieshare-gin/forms"
	"github.com/VieShare/vieshare-gin/routers"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/joho/godotenv"
)


// @title           VieShare Gin Private API
// @version         2.0
// @description     PocketBase-compatible REST API for VieShare e-commerce platform with SQLite, Redis and JWT authentication
// @termsOfService  https://vieshare.com/terms

// @contact.name   VieShare API Support
// @contact.url    https://vieshare.com/support
// @contact.email  support@vieshare.com

// @license.name  MIT License
// @license.url   https://github.com/VieShare/vieshare-gin/blob/master/LICENSE

// @host      localhost:9000
// @BasePath  /api

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	//Load the .env file
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("error: failed to load the env file")
	}

	if os.Getenv("ENV") == "PRODUCTION" {
		gin.SetMode(gin.ReleaseMode)
	}

	//Start the default gin server
	r := gin.Default()

	//Custom form validator
	binding.Validator = new(forms.DefaultValidator)

	// Setup middlewares
	r.Use(routers.CORSMiddleware())
	r.Use(routers.RequestIDMiddleware())
	r.Use(gzip.Gzip(gzip.DefaultCompression))

	//Start SQLite3 database
	//Example: db.GetDB() - More info in the models folder
	db.Init()

	//Start Redis on database 1 - it's used to store the JWT but you can use it for anythig else
	//Example: db.GetRedis().Set(KEY, VALUE, at.Sub(now)).Err()
	db.InitRedis(1)

	// Setup V1 API routes
	v1 := r.Group("/v1")
	routers.SetupV1Routes(v1, routers.TokenAuthMiddleware())

	// Setup PocketBase-compatible API routes
	api := r.Group("/api")
	routers.SetupPocketBaseRoutes(api)

	// Setup static routes and documentation
	routers.SetupStaticRoutes(r)

	port := os.Getenv("PORT")

	log.Printf("\n\n PORT: %s \n ENV: %s \n SSL: %s \n Version: %s \n\n", port, os.Getenv("ENV"), os.Getenv("SSL"), os.Getenv("API_VERSION"))

	if os.Getenv("SSL") == "TRUE" {

		//Generated using sh generate-certificate.sh
		SSLKeys := &struct {
			CERT string
			KEY  string
		}{
			CERT: "./cert/myCA.cer",
			KEY:  "./cert/myCA.key",
		}

		r.RunTLS(":"+port, SSLKeys.CERT, SSLKeys.KEY)
	} else {
		r.Run(":" + port)
	}

}
