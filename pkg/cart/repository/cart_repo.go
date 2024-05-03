package repository

import "github.com/Clareand/web-chart/pkg/cart/model"

type CartRepo interface {
	GetCart(customer_id string) ([]model.Carts, error)
	AddToCart(param model.AddToCart) error
}
