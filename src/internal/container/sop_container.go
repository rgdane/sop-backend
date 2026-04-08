package container

import (
	"jk-api/api/http/controllers/v1/handlers"
	"jk-api/internal/repository/graphdb"
	"jk-api/internal/repository/sql"
	"jk-api/internal/service"
)

func InitSopContainer() *handlers.SopHandler {
	repo := sql.NewSopRepository()
	graphRepo := graphdb.NewSopRepository()
	service := service.NewSopService(repo, graphRepo)
	return handlers.NewSopHandler(service)
}
