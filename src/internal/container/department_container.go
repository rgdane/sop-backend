package container

import (
	"jk-api/api/http/controllers/v1/handlers"
	"jk-api/internal/repository/sql"
	"jk-api/internal/service"
)

func InitDepartmentContainer() *handlers.DepartmentHandler {
	repo := sql.NewDepartmentRepository()
	service := service.NewDepartmentService(repo)
	return handlers.NewDepartmentHandler(service)
}
