package router

import (
	"strings"

	"github.com/gin-gonic/gin"
	clients_fleet "github.com/tariq-ventura/logistic-service/internal/clients/fleet"
	"github.com/tariq-ventura/logistic-service/internal/interfaces"
	"github.com/tariq-ventura/logistic-service/internal/logging"
	requets_db "github.com/tariq-ventura/logistic-service/internal/requests/db"
	"github.com/tariq-ventura/logistic-service/internal/validations"
)

type Routes struct {
	Routes      *gin.Engine
	Context     *gin.Context
	FleetClient *clients_fleet.Client
	Logging     logging.ILogging
	Trace       interfaces.ITrace
	RequestsDB  requets_db.IRequestsDB
}

func (r *Routes) SetupRouter() *gin.Engine {
	r.Routes = gin.Default()

	r.SetupCors()

	r.HealthCheckRoutes()
	r.RequestsRoutes(r.Routes)
	return r.Routes
}

func (r *Routes) Run() {
	const defaultPort = "3001"

	port, err := validations.RequiredEnv("PORT")

	if err != nil {
		port = defaultPort

		r.Logging.LogInfo(
			"La variable PORT no está definida; el servidor se iniciará en el puerto predeterminado",
			map[string]any{
				"port": defaultPort,
			},
		)
	}

	if !strings.HasPrefix(port, ":") {
		port = ":" + port
	}

	r.Logging.LogInfo(
		"Iniciando servidor HTTP",
		map[string]any{
			"address": port,
		},
	)

	if err := r.Routes.Run(port); err != nil {
		r.Logging.LogError(
			"No se pudo iniciar el servidor HTTP",
			map[string]any{
				"address": port,
				"error":   err.Error(),
			},
		)
	}
}
