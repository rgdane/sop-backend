package container

import (
	"jk-api/api/http/controllers/v1/handlers"
	"jk-api/internal/repository/sql"
	"jk-api/internal/service"
)

func InitSopContainer() *handlers.SopHandler {
	repo := sql.NewSopRepository()
	service := service.NewSopService(repo)
	return handlers.NewSopHandler(service)
}
