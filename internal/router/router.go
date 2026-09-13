package router

import (
	"strings"

	"github.com/gin-gonic/gin"
	equipments_db "github.com/tariq-ventura/logistic-service/internal/equipments/db"
	"github.com/tariq-ventura/logistic-service/internal/interfaces"
	"github.com/tariq-ventura/logistic-service/internal/logging"
	requests_db "github.com/tariq-ventura/logistic-service/internal/requests/db"
	"github.com/tariq-ventura/logistic-service/internal/validations"
)

type Routes struct {
	Routes       *gin.Engine
	Context      *gin.Context
	Logging      logging.ILogging
	Trace        interfaces.ITrace
	EquipmentsDB equipments_db.IEquipmentsDB
	RequestsDB   requests_db.IRequestsDB
}

func (r *Routes) SetupRouter() *gin.Engine {
	r.Routes = gin.Default()
	r.SetupCors()
	r.HealthCheckRoutes()
	r.EquipmentsRoutes(r.Routes)
	r.RequestsRoutes(r.Routes)
	return r.Routes
}
func (r *Routes) Run() {
	const defaultPort = "3001"
	port, err := validations.RequiredEnv("PORT")
	if err != nil {
		port = defaultPort
		r.Logging.LogInfo("La variable PORT no está definida; se usará el puerto predeterminado", map[string]any{"port": defaultPort})
	}
	if !strings.HasPrefix(port, ":") {
		port = ":" + port
	}
	if err := r.Routes.Run(port); err != nil {
		r.Logging.LogError("No se pudo iniciar el servidor HTTP", map[string]any{"address": port, "error": err.Error()})
	}
}
