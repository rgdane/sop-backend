package container

import (
	"jk-api/api/http/controllers/v1/handlers"
	"jk-api/internal/repository/graphdb"
	"jk-api/internal/repository/sql"
	"jk-api/internal/service"
)

func InitSpkContainer() *handlers.SpkHandler {
	repo := sql.NewSpkRepository()
	graphRepo := graphdb.NewSpkRepository()
	service := service.NewSpkService(repo, graphRepo)
	return handlers.NewSpkHandler(service)
}
