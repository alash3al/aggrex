package server

import (
	"core.globals"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

// Serve start the restful server
func Serve(addr string) error {
	e := echo.New()
	e.HideBanner = true

	e.Pre(middleware.RemoveTrailingSlash())
	e.Use(middleware.Recover())
	e.Use(middleware.Logger())
	e.Use(middleware.CORS())
	e.Use(middleware.BodyLimit(*globals.FlagBodyMaxSize))
	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{
		Level: 9,
	}))

	e.GET("/", hello)
	e.POST("/", quickExec, MustBeAdminMiddleware())
	e.POST("/procedure/search", procedureSearch, MustBeAdminMiddleware())
	e.POST("/procedure/:key", procedureSave, MustBeAdminMiddleware())
	e.DELETE("/procedure/:key", procedureDelete, MustBeAdminMiddleware())
	e.GET("/procedure/:key/result", procedureExec)
	e.POST("/procedure/:key/result", procedureExec)
	e.POST("/globals/vars", globalsSetVars, MustBeAdminMiddleware())
	e.POST("/globals/var/:key", globalsSetVar, MustBeAdminMiddleware())
	e.GET("/globals/vars", globalsGetVars, MustBeAdminMiddleware())

	return e.Start(addr)
}
