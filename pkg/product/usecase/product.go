package usecase

import "github.com/Clareand/web-chart/pkg/product/model"

type ProductUsecase interface {
	GetProductList() ([]model.Products, error)
}
