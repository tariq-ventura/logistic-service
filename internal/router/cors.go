package router

import "github.com/gin-contrib/cors"

func (r *Routes) SetupCors() {
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:8080"}
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization", "Access-Control-Allow-Origin"}
	config.ExposeHeaders = []string{"Content-Length"}
	config.AllowCredentials = true
	config.MaxAge = 12 * 60 * 60
	r.Routes.Use(cors.New(config))
}
