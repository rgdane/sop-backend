package container

import (
	"jk-api/api/http/controllers/v1/handlers"
	"jk-api/internal/repository/sql"
	"jk-api/internal/service"
)

func InitDatabaseNodeContainer() *handlers.DatabaseNodeHandler {
	repo := sql.NewDatabaseNodeRepository()
	service := service.NewDatabaseNodeService(repo)
	handler := handlers.NewDatabaseNodeHandler(service)

	return handler
}
