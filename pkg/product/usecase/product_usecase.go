package usecase

import (
	"github.com/Clareand/web-chart/pkg/product/model"
	"github.com/Clareand/web-chart/pkg/product/repository"
)

type productUsecase struct {
	repo repository.ProductRepo
}

func NewProductUsecase(repo repository.ProductRepo) ProductUsecase {
	return &productUsecase{repo: repo}
}

func (u *productUsecase) GetProductList() ([]model.Products, error) {
	return u.repo.GetProduct()
}
