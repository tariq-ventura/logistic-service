package router

import (
	"github.com/gin-gonic/gin"
	requests_handlers "github.com/tariq-ventura/logistic-service/internal/requests/handlers"
)

func (ro *Routes) RequestsRoutes(r *gin.Engine) {
	h := requests_handlers.NewRequestHandler(ro.Context, ro.RequestsDB, ro.Trace, ro.Logging)
	routes := r.Group("/api/v1/requests")
	{
		routes.POST("", h.CreateRequest)
		routes.GET("", h.ListRequests)
		routes.POST("/search", h.SearchRequests)
		routes.GET("/:requestID", h.ListRequestById)
		routes.PATCH("/:requestID", h.UpdateRequest)
		routes.DELETE("/:requestID", h.RemoveRequest)
		routes.PATCH("/:requestID/status", h.UpdateRequestStatus)
		routes.GET("/:requestID/status-history", h.ListRequestStatusHistory)
		routes.PATCH("/:requestID/assignment", h.AssignMachinery)
		routes.DELETE("/:requestID/assignment", h.ReleaseMachinery)
	}
}
