package requests_handlers

import (
	"github.com/gin-gonic/gin"
	clients_fleet "github.com/tariq-ventura/logistic-service/internal/clients/fleet"
	"github.com/tariq-ventura/logistic-service/internal/interfaces"
	"github.com/tariq-ventura/logistic-service/internal/logging"
	requets_db "github.com/tariq-ventura/logistic-service/internal/requests/db"
	requests_domain "github.com/tariq-ventura/logistic-service/internal/requests/domain"
)

type RequestHandler struct {
	db          requets_db.IRequestsDB
	fleetClient *clients_fleet.Client
	trace       interfaces.ITrace
	logs        logging.ILogging
}

func NewRequestHandler(server *gin.Context, db requets_db.IRequestsDB, trace interfaces.ITrace, logs logging.ILogging, fleetClient *clients_fleet.Client) requests_domain.IRequests {
	return &RequestHandler{
		db:          db,
		fleetClient: fleetClient,
		trace:       trace,
		logs:        logs,
	}
}
