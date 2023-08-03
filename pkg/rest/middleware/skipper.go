package middleware

import (
	"github.com/labstack/echo/v4"
)

var paths2ignore = []string{"/health", "/metrics"}

func Skipper(c echo.Context) bool {
	for _, path := range paths2ignore {
		if path == c.Path() {
			return true
		}
	}
	return false
}
