package requests_handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/tariq-ventura/logistic-service/internal/interfaces"
	"github.com/tariq-ventura/logistic-service/internal/logging"
	requests_db "github.com/tariq-ventura/logistic-service/internal/requests/db"
	requests_domain "github.com/tariq-ventura/logistic-service/internal/requests/domain"
)

type RequestHandler struct {
	db    requests_db.IRequestsDB
	trace interfaces.ITrace
	logs  logging.ILogging
}

func NewRequestHandler(_ *gin.Context, db requests_db.IRequestsDB, trace interfaces.ITrace, logs logging.ILogging) requests_domain.IRequests {
	return &RequestHandler{db: db, trace: trace, logs: logs}
}
