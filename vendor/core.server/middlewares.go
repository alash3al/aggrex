package server

import (
	"errors"

	"core.globals"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

// MustBeAdminMiddleware .
func MustBeAdminMiddleware() echo.MiddlewareFunc {
	return middleware.KeyAuth(func(key string, c echo.Context) (ok bool, err error) {
		ok = key == *globals.FlagAdminToken
		if !ok {
			err = errors.New("Invalid Admin token specified")
		}
		return ok, err
	})
}
