package container

import (
	"jk-api/api/http/controllers/v1/handlers"
	"jk-api/internal/repository/graphdb"
	"jk-api/internal/repository/sql"
	"jk-api/internal/service"
)

func InitDashboardContainer() *handlers.DashboardHandler {
	repo := sql.NewDashboardRepository()
	graphRepo := graphdb.NewDashboardRepository()
	service := service.NewDashboardService(repo, graphRepo)
	return handlers.NewDashboardHandler(service)
}
