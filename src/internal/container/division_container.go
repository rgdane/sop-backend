package container

import (
	"jk-api/api/http/controllers/v1/handlers"
	"jk-api/internal/repository/sql"
	"jk-api/internal/service"
)

func InitDivisionContainer() *handlers.DivisionHandler {
	repo := sql.NewDivisionRepository()
	service := service.NewDivisionService(repo)
	return handlers.NewDivisionHandler(service)
}
