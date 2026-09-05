package assignments_handlers

import (
	"github.com/gin-gonic/gin"
	assignmets_db "github.com/tariq-ventura/logistic-service/internal/assigments/db"
	assignments_domain "github.com/tariq-ventura/logistic-service/internal/assigments/domain"
	clients_fleet "github.com/tariq-ventura/logistic-service/internal/clients/fleet"
	"github.com/tariq-ventura/logistic-service/internal/interfaces"
	"github.com/tariq-ventura/logistic-service/internal/logging"
	requets_db "github.com/tariq-ventura/logistic-service/internal/requests/db"
)

type AssignmentHandler struct {
	db          assignmets_db.IAssignmentsDB
	requestsDB  requets_db.IRequestsDB
	fleetClient clients_fleet.Client
	trace       interfaces.ITrace
	logs        logging.ILogging
}

func NewAssignmentHandler(
	server *gin.Context,
	db assignmets_db.IAssignmentsDB,
	trace interfaces.ITrace,
	logs logging.ILogging,
	requestsDB requets_db.IRequestsDB,
	fleetClient clients_fleet.Client,
) assignments_domain.IAssignments {
	return &AssignmentHandler{
		db:          db,
		requestsDB:  requestsDB,
		trace:       trace,
		logs:        logs,
		fleetClient: fleetClient,
	}
}
