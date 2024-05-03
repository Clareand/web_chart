package repository

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/Clareand/web-chart/config/postgresql"
	_redis "github.com/Clareand/web-chart/config/redis"
	"github.com/Clareand/web-chart/libs/models"
	"github.com/Clareand/web-chart/pkg/auth/model"
	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
	"go.elastic.co/apm/v2"
	"golang.org/x/crypto/bcrypt"
)

type loginRepo struct {
	dbConn    *postgresql.DbConnection
	ctx       context.Context
	echoCtx   echo.Context
	redisConn _redis.RedisConnection
}

type hashLogin struct {
	Password   string `json:"password"`
	CustomerID string `json:"customer_id"`
}

func NewLoginRepo(dbConn *postgresql.DbConnection, redisConn _redis.RedisConnection) LoginRepo {
	return &loginRepo{dbConn: dbConn, redisConn: redisConn}
}

func (r *loginRepo) WithContext(echoCtx echo.Context) LoginRepo {
	r.ctx = echoCtx.Request().Context()
	r.dbConn.Db = r.dbConn.Db.WithContext(r.ctx)
	r.echoCtx = echoCtx
	return r
}

const (
	DDMMYYYYhhmmss = "2006-01-02 15:04:05.000"

	MAP_CONFIG_EXPIRY = 1 * time.Hour * 24

	RedisKeySession = "sessions"
)

func (r *loginRepo) CheckUser(username string) model.CheckUserIsTrue {
	span, _ := apm.StartSpan(r.ctx, "CheckUser", "repository")
	defer span.End()

	result := model.CheckUserIsTrue{}

	q := r.dbConn.Db

	sql := "SELECT * FROM public.f_check_user(?)"
	q.Raw(sql, username).Scan(&result)
	return result
}

func (r *loginRepo) GetPassword(userId string) (password string) {
	span, _ := apm.StartSpan(r.ctx, "public.f_get_password_user", "repository")
	defer span.End()

	result := model.DataUserFromDB{}

	q := r.dbConn.Db
	q.Raw("SELECT * FROM public.f_get_password_user(?)", userId).Scan(&result)

	return result.Credential
}

func (r *loginRepo) GetDataUser(userID string, rememberMe bool) model.GetterUserLogin {
	span, _ := apm.StartSpan(r.ctx, "GetDataUser", "repository")
	defer span.End()

	result := model.GetterUserLogin{}
	session := model.GetterUserLogin{}

	q := r.dbConn.Db

	var intervalMonthRemember, intervalRememberDay string
	intervalRemember := os.Getenv("REMEMBER_ME")
	if intervalRemember == "" {
		intervalMonthRemember = "'1 month'"
	} else {
		intervalMonthRemember = fmt.Sprintf(`'%v'`, intervalRemember)
	}
	intervalMonthRemember = strings.ReplaceAll(intervalMonthRemember, "'", "")

	defaultInterval := os.Getenv("DEFAULT_REMEMBER")
	if defaultInterval == "" {
		intervalRememberDay = "'1 day'"
	} else {
		intervalRememberDay = fmt.Sprintf(`'%v'`, defaultInterval)
	}

	intervalRememberDay = strings.ReplaceAll(intervalRememberDay, "'", "")
	fmt.Println("Remember Me : ", rememberMe)
	fmt.Println("Interval Day Remember : ", intervalRememberDay)
	fmt.Println("Interval Month Remember : ", intervalMonthRemember)

	qr := "SELECT * FROM public.f_get_log_data_user(?)"
	q.Raw(qr, userID).Scan(&result)

	qrCreateRefreshToken := "SELECT * FROM public.f_create_session_refresh_token_user(?,?,?,?,?)"
	q.Raw(qrCreateRefreshToken, userID, rememberMe, intervalMonthRemember, intervalRememberDay, nil).Scan(&session)

	/* START CREATE SESSION CACHE REDIS */
	// recheck
	var sessionData model.Session
	q.Raw("select * from public.sessions s where (encode(public.hmac(s.id::text, s.key, 'sha256'), 'hex')) = ?", session.RefreshToken).Scan(&sessionData)

	fmt.Println(sessionData, "session")

	r.redisConn.CreateCache(r.echoCtx, RedisKeySession, RedisKeySession+"_"+session.RefreshToken, sessionData, MAP_CONFIG_EXPIRY)
	/* END CREATE SESSION CACHE REDIS */

	result.RefreshToken = session.RefreshToken

	fmt.Println("user on ", result)

	return result
}
func (r *loginRepo) CreateToken(getLogin model.GetterUserLogin) (token string) {
	user := model.User{
		CustomerID: getLogin.CustomerID,
		UserName:   getLogin.UserName,
		UserEmail:  getLogin.UserEmail,
		CreatedAt:  getLogin.CreatedAt,
		Session:    getLogin.Session,
	}

	token, exp := createTokenUserFunc(user, getLogin, r.echoCtx)
	tm := time.Unix(exp, 0)
	fmt.Println("Time Expired : ", tm)

	return token
}

func (r *loginRepo) ResultLogin(param model.GetterUserLogin, token string) model.ResultLoginUserBE {
	span, _ := apm.StartSpan(r.ctx, "ResultLogin", "repository")
	defer span.End()

	getLogin := param

	user := model.User{
		CustomerID: getLogin.CustomerID,
		UserName:   getLogin.UserName,
		UserEmail:  getLogin.UserEmail,
		CreatedAt:  getLogin.CreatedAt,
		Session:    getLogin.Session,
	}
	accessToken := model.AccessToken{
		Type:         "bearer",
		Token:        token,
		RefreshToken: getLogin.RefreshToken,
	}

	result := model.ResultLoginUserBE{
		User:        user,
		AccessToken: accessToken,
	}

	return result
}

func (r *loginRepo) CheckPasswordHash(password string, hash string) bool {

	fmt.Println("CheckPasswordHash pw", password)
	fmt.Println("CheckPasswordHash hash", hash)
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	fmt.Println("CheckPasswordHash err", err)

	return err == nil
}

func hashAndSalt(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)

	if err != nil {
		log.Println(err)
	}
	return string(bytes), err
}

func (r *loginRepo) RefreshToken(req model.AccessToken) <-chan models.Result {
	output := make(chan models.Result)
	var getLogin model.GetterUserLogin
	var checkSession model.CheckSession

	go func() {
		span, _ := apm.StartSpan(r.ctx, "RefreshToken", "repository")
		defer span.End()
		defer close(output)
		q := r.dbConn.Db
		qr := "SELECT * FROM public.f_check_session_refresh_token(?)"
		q.Raw(qr, req.RefreshToken).Scan(&checkSession)

		if !checkSession.IsNotExpired {
			/* START DELETE SESSION CACHE REDIS */
			r.redisConn.DeleteCache(r.echoCtx, RedisKeySession, RedisKeySession+"_"+req.RefreshToken)
			/* END DELETE SESSION CACHE REDIS */
			resp := &models.Response{Code: 400, MessageCode: 0000, Message: "Session Expired"}
			output <- models.Result{Data: resp}
		} else if checkSession.CustomerID != "" {

			var intervalMonthRemember, intervalRememberDay string
			intervalRemember := os.Getenv("REMEMBER_ME")
			if intervalRemember == "" {
				intervalMonthRemember = "'1 month'"
			} else {
				intervalMonthRemember = fmt.Sprintf(`'%v'`, intervalRemember)
			}

			defaultInterval := os.Getenv("DEFAULT_REMEMBER")
			if defaultInterval == "" {
				intervalRememberDay = "'1 day'"
			} else {
				intervalRememberDay = fmt.Sprintf(`'%v'`, defaultInterval)
			}

			var paramTimeUpdateSession string
			if checkSession.RememberMe {
				paramTimeUpdateSession = intervalMonthRemember
			} else {
				paramTimeUpdateSession = intervalRememberDay
			}

			q.Exec(`SELECT public.f_update_session(?,?)`, req.RefreshToken, paramTimeUpdateSession)

			/* START UPDATE SESSION CACHE REDIS */
			var sessionData model.Session
			q.Raw("select * from public.sessions s where (encode(public.hmac(s.id::text, s.key, 'sha256'), 'hex')) = ?", req.RefreshToken).Scan(&sessionData)

			r.redisConn.CreateCache(r.echoCtx, RedisKeySession, RedisKeySession+"_"+req.RefreshToken, sessionData, MAP_CONFIG_EXPIRY)
			/* END UPDATE SESSION CACHE REDIS */

			getLogin.RefreshToken = req.RefreshToken
			if getLogin.CustomerID != "" {
				user := model.User{
					CustomerID: getLogin.CustomerID,
					UserName:   getLogin.UserName,
					UserEmail:  getLogin.UserEmail,
					CreatedAt:  getLogin.CreatedAt,
					Session:    getLogin.Session,
				}
				token, exp := createTokenUserFunc(user, getLogin, r.echoCtx)

				tm := time.Unix(exp, 0)
				fmt.Println(tm)
				accessToken := model.AccessToken{
					Type:         "bearer",
					Token:        token,
					RefreshToken: req.RefreshToken,
				}

				result := model.ResultLoginUserBE{
					User:        user,
					AccessToken: accessToken,
				}

				data := &result
				resp := &models.Response{Code: 200, MessageCode: 0000, Message: "successfully", Data: data}
				output <- models.Result{Data: resp}
			}

			resp := &models.Response{Code: 400, MessageCode: 0000, Message: "Refresh Failed"}
			output <- models.Result{Data: resp}
		}
	}()

	return output
}

func createTokenUserFunc(user model.User, getLogin model.GetterUserLogin, c echo.Context) (token string, exp int64) {

	// env configure token exp
	var duration time.Duration
	timeDuration, err := time.ParseDuration(os.Getenv("LIMIT_TOKEN"))
	if err != nil {
		duration = 24 * time.Hour
	} else {
		duration = timeDuration
	}
	// github.com/dgrijalva/jwt-go
	// github.com/dgrijalva/jwt-go
	// Create token
	claims := jwt.MapClaims{}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Set claims
	claims["user"] = user
	claims["refresh_token"] = getLogin.RefreshToken
	claims["session"] = getLogin.Session
	exp_ := time.Now().Add(duration).Unix()
	claims["exp"] = exp_
	// claims["exp"] = time.Now().Add(time.Hour * 24).Unix()
	claims["iat"] = time.Now().Unix()

	// Generate encoded token and send it as response.
	token, _ = tok.SignedString([]byte(os.Getenv("API_SECRET")))
	return token, exp_
}

func (r *loginRepo) Logout(req model.AccessToken) <-chan models.Result {
	output := make(chan models.Result)
	var responseProcedure model.ResponseDeleteSessionRefreshToken

	go func() {
		q := r.dbConn.Db
		qr := "SELECT * FROM system_configuration.p_delete_session(?)"
		q.Raw(qr, req.RefreshToken).Scan(&responseProcedure)

		if !responseProcedure.IsSuccess {
			/* START DELETE SESSION CACHE REDIS */
			r.redisConn.DeleteCache(r.echoCtx, RedisKeySession, RedisKeySession+"_"+req.RefreshToken)
			/* END DELETE SESSION CACHE REDIS */

			resp := &models.Response{Code: 200, MessageCode: 0000, Message: "successfully"}
			output <- models.Result{Data: resp}
		} else {
			resp := &models.Response{Code: 400, MessageCode: 0000, Message: "Logout Failed"}
			output <- models.Result{Data: resp}
		}
	}()

	return output
}
