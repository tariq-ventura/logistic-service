package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (r *Routes) HealthCheckRoutes() {
	r.Routes.GET("/health", r.print)
}

func (r *Routes) print(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"Status": "Up and Running"})
}
