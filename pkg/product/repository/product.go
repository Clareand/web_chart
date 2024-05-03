package repository

import (
	"context"

	"github.com/Clareand/web-chart/config/postgresql"
	"github.com/Clareand/web-chart/pkg/product/model"
	"github.com/labstack/echo/v4"
)

type productRepo struct {
	dbConn  *postgresql.DbConnection
	ctx     context.Context
	echoCtx echo.Context
}

func NewProductRepo(dbConn *postgresql.DbConnection) ProductRepo {
	return &productRepo{dbConn: dbConn}
}

func (r *productRepo) GetProduct() ([]model.Products, error) {
	var products []model.Products

	sql := `select * from public.f_get_all_product()`
	err := r.dbConn.Db.Raw(sql).Scan(&products).Error

	if err != nil {
		return nil, err
	}
	return products, nil
}
