package handler

import (
	"github.com/Clareand/web-chart/libs/models"
	"github.com/labstack/echo/v4"
)

func (h *HttpHandler) GetProductList(c echo.Context) error {
	data, err := h.usecase.GetProductList()
	if err != nil {
		return models.ToJSON(c).InternalServerError(err.Error())
	}
	return models.ToJSON(c).Ok(data, "Successfully")
}
