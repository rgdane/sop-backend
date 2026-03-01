package container

import (
	"jk-api/api/http/controllers/v1/handlers"
)

type AppContainer struct {
	AuthHandler               *handlers.AuthHandler
	DepartmentHandler         *handlers.DepartmentHandler
	DivisionHandler           *handlers.DivisionHandler
	LevelHandler              *handlers.LevelHandler
	PermissionHandler         *handlers.PermissionHandler
	PositionHandler           *handlers.PositionHandler
	RoleHandler               *handlers.RoleHandler
	TitleHandler              *handlers.TitleHandler
	UserHandler               *handlers.UserHandler
	SopHandler                *handlers.SopHandler
	SpkHandler                *handlers.SpkHandler
	SpkJobHandler             *handlers.SpkJobHandler
	SopJobHandler             *handlers.SopJobHandler
	DatabaseNodeHandler       *handlers.DatabaseNodeHandler
}

func NewAppContainer() *AppContainer {
	return &AppContainer{
		AuthHandler:               InitAuthContainer(),
		DivisionHandler:           InitDivisionContainer(),
		LevelHandler:              InitLevelContainer(),
		PermissionHandler:         InitPermissionContainer(),
		PositionHandler:           InitPositionContainer(),
		RoleHandler:               InitRoleContainer(),
		TitleHandler:              InitTitleContainer(),
		UserHandler:               InitUserContainer(),
		SopHandler:                InitSopContainer(),
		SpkHandler:                InitSpkContainer(),
		SpkJobHandler:             InitSpkJobContainer(),
		SopJobHandler:             InitSopJobContainer(),
		DatabaseNodeHandler:       InitDatabaseNodeContainer(),
	}
}
