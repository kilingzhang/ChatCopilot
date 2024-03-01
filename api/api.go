package api

import (
	"fmt"

	"github.com/lw396/WeComCopilot/service"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Api struct {
	port    int
	service *service.Service
}

type Config struct {
	App  *service.Service
	Port int
}

func New(c Config) *Api {
	return &Api{
		port:    c.Port,
		service: c.App,
	}
}

func (api *Api) Run() error {
	engine := echo.New()
	engine.HTTPErrorHandler = HTTPErrorHandler
	engine.Validator = NewValidator()

	engine.Use(middleware.CORS())
	engine.Use(middleware.Recover())
	engine.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{echo.GET, echo.PUT, echo.POST},
	}))

	v1 := engine.Group("/v1")

	v1.GET("/group_contact", api.getGroupContact)
	v1.GET("/group_message", api.getGroupMessage)

	_ = v1.Group("/user", api.Authenticate)
	{

	}

	return engine.Start(fmt.Sprintf(":%d", api.port))
}
