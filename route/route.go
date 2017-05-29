package route

import (
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"gophass_server/controller/api"
	"net/http"
)

func Routing() *echo.Echo {
	e := echo.New()

	e.Use(middleware.Logger())

	apiv1 := e.Group("/api/v1")
	apiv1.GET("/events/search", api.GetSearchEvent)

	return e
}
