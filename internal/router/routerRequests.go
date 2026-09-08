package router

import (
	"github.com/gin-gonic/gin"
	requests_handlers "github.com/tariq-ventura/logistic-service/internal/requests/handlers"
)

func (ro *Routes) RequestsRoutes(r *gin.Engine) {
	rr := requests_handlers.NewRequestHandler(ro.Context, ro.RequestsDB, ro.Trace, ro.Logging, ro.FleetClient, ro.AssignmentsDB)
	routes := r.Group("/api/v1/requests")
	{
		routes.POST("", rr.CreateRequest)
		routes.GET("", rr.ListRequests)
		routes.Handle("QUERY", "", rr.SearchRequests)
		routes.GET("/:requestID", rr.ListRequestById)
		routes.PATCH("/:requestID", rr.UpdateRequest)
		routes.PATCH("/:requestID/status", rr.UpdateRequestStatus)
		routes.GET("/:requestID/status-history", rr.ListRequestStatusHistory)
		routes.GET("/:requestID/recommendations", rr.ListRequestRecommendations)
		routes.GET("/:requestID/assignment", rr.ListAssginmentByRequest)
		routes.POST("/:requestID/assignment", rr.CreateAssignment)
	}
}
