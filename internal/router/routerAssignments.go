package router

import (
	"github.com/gin-gonic/gin"
	assignments_handlers "github.com/tariq-ventura/logistic-service/internal/assigments/handlers"
)

func (ro *Routes) AssignmentsRoutes(r *gin.Engine) {
	ar := assignments_handlers.NewAssignmentHandler(ro.Context, ro.AssignmentsDB, ro.Trace, ro.Logging, ro.RequestsDB, *ro.FleetClient)
	routes := r.Group("/api/v1/assignments")
	{
		routes.GET("", ar.ListAssginments)
		routes.GET("/:assignmentID", ar.ListAssginmentById)
		routes.PATCH("/:assignmentID/status", ar.UpdateAssignmentStatus)
	}
}
