package container

import (
	"jk-api/api/http/controllers/v1/handlers"
	"jk-api/internal/repository/sql"
	"jk-api/internal/service"
)

func InitSpkJobContainer() *handlers.SpkJobHandler {
	repo := sql.NewSpkJobRepository()
	service := service.NewSpkJobService(repo)
	return handlers.NewSpkJobHandler(service)
}
