package api

import (
	"github.com/labstack/echo"
	"gophass_server/model/events"
	"net/http"
)

// GetSearchEvent : GET /api/v1/events/search
func GetSearchEvent(c echo.Context) (err error) {
	keywords := c.QueryParam("keywords")
	month := c.QueryParam("month")
	address := c.QueryParam("address")

	seachResult := events.SearchEvents()

	return c.JSON(http.StatusOK, seachResult)
}
