package main

import (
	"jk-api/cmd/bootstrap"
)

//	@title			JK API
//	@version		1.0
//	@description	JalanKerja API Server
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	API Support
//	@contact.url	http://www.swagger.io/support
//	@contact.email	support@swagger.io

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

//	@BasePath	/api/v1

// @securityDefinitions.apikey	BearerAuth
// @in							header
// @name						Authorization
// @description					JWT Authorization header using the Bearer scheme. Example: "Bearer {token}"
func main() {
	bootstrap.Setup()
}
