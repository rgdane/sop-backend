package container

import (
	"jk-api/api/http/controllers/v1/handlers"
)

type AppContainer struct {
	AuthHandler               *handlers.AuthHandler
	DivisionHandler           *handlers.DivisionHandler
	PermissionHandler         *handlers.PermissionHandler
	RoleHandler               *handlers.RoleHandler
	TitleHandler              *handlers.TitleHandler
	UserHandler               *handlers.UserHandler
	SopHandler                *handlers.SopHandler
	SpkHandler                *handlers.SpkHandler
	SpkJobHandler             *handlers.SpkJobHandler
	SopJobHandler             *handlers.SopJobHandler
	DashboardHandler		  *handlers.DashboardHandler
}

func NewAppContainer() *AppContainer {
	return &AppContainer{
		AuthHandler:               InitAuthContainer(),
		DivisionHandler:           InitDivisionContainer(),
		PermissionHandler:         InitPermissionContainer(),
		RoleHandler:               InitRoleContainer(),
		TitleHandler:              InitTitleContainer(),
		UserHandler:               InitUserContainer(),
		SopHandler:                InitSopContainer(),
		SpkHandler:                InitSpkContainer(),
		SpkJobHandler:             InitSpkJobContainer(),
		SopJobHandler:             InitSopJobContainer(),
		DashboardHandler:          InitDashboardContainer(),
	}
}
