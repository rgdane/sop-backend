package container

import (
	"jk-api/api/http/controllers/v1/handlers"
	"jk-api/internal/repository/sql"
	"jk-api/internal/service"
)

func InitAuthContainer() *handlers.AuthHandler {
	repo := sql.NewUserRepository()
	service := service.NewAuthService(repo)
	return handlers.NewAuthHandler(service)
}
