package middleware

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v4/middleware"
)

func RequestIDConfig() middleware.RequestIDConfig {
	return middleware.RequestIDConfig{
		Skipper: middleware.DefaultRequestIDConfig.Skipper,
		Generator: func() string {
			return uuid.NewString()
		},
		TargetHeader: middleware.DefaultRequestIDConfig.TargetHeader,
	}
}
