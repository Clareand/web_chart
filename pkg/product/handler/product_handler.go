package handler

import (
	"github.com/Clareand/web-chart/config/postgresql"
	"github.com/Clareand/web-chart/pkg/product/usecase"
	"github.com/labstack/echo/v4"
)

type HttpHandler struct {
	usecase usecase.ProductUsecase
}

func NewHTTPHandler(usecase usecase.ProductUsecase) *HttpHandler {
	return &HttpHandler{usecase: usecase}
}

func (h *HttpHandler) Mount(g *echo.Group, auth echo.MiddlewareFunc, dbConn *postgresql.DbConnection) {
	g.GET("/product", h.GetProductList, auth)
}
