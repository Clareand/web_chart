package repository

import (
	"github.com/Clareand/web-chart/libs/models"
	"github.com/Clareand/web-chart/pkg/auth/model"
	"github.com/labstack/echo/v4"
)

type LoginRepo interface {
	Logout(req model.AccessToken) <-chan models.Result
	RefreshToken(req model.AccessToken) <-chan models.Result
	// check user di DB
	CheckUser(username string) model.CheckUserIsTrue
	// check password di DB
	GetPassword(userId string) (password string)
	// check password antara Param dari DB
	CheckPasswordHash(password string, hash string) bool
	// menampilkan data dari user
	GetDataUser(userID string, rememberMe bool) model.GetterUserLogin
	// create token
	CreateToken(req model.GetterUserLogin) (token string)
	//Login Result
	ResultLogin(param model.GetterUserLogin, token string) model.ResultLoginUserBE
	//monitoring span
	WithContext(echo.Context) LoginRepo
}
