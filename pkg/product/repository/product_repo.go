package repository

import "github.com/Clareand/web-chart/pkg/product/model"

type ProductRepo interface {
	GetProduct() ([]model.Products, error)
}
