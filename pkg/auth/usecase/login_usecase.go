package usecase

import (
	"github.com/Clareand/web-chart/libs/models"
	"github.com/Clareand/web-chart/pkg/auth/model"
	"github.com/labstack/echo/v4"
)

type LoginUsecase interface {
	NewLoginUser(req model.ReqNewLogin, ipnumber string) <-chan models.Result
	RefreshToken(req model.AccessToken) <-chan models.Result
	Logout(req model.AccessToken) <-chan models.Result
	//monitoring span
	WithContext(echo.Context) LoginUsecase
}
