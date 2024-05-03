package handler

import (
	"io/ioutil"

	"github.com/Clareand/web-chart/libs/models"
	"github.com/labstack/echo/v4"
)

func (h *HttpHandler) GetCart(c echo.Context) error {
	data, err := h.usecase.GetCart()
	if err != nil {
		return models.ToJSON(c).InternalServerError(err.Error())
	}
	return models.ToJSON(c).Ok(data, "Successfully")
}

func (h *HttpHandler) AddToCart(c echo.Context) error {
	body, err := ioutil.ReadAll(c.Request().Body)

	if err != nil {
		return models.ToJSON(c).BadRequest("Bad Request")
	}
}
