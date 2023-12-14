package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"log"
	"net/http"
	"strings"
	"time"
	"webook/internal/web"
)

type LoginJWTMiddlewareBuilder struct {
}

func (m *LoginJWTMiddlewareBuilder) CheckLogin() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		path := ctx.Request.URL.Path
		if path == "/users/signup" ||
			path == "/users/login" ||
			path == "/users/login_sms/code/send" ||
			path == "/users/login_sms" {
			return
		}
		//根据约定token在Authorization头部
		authCode := ctx.GetHeader("Authorization")
		//没登录，没有token
		if authCode == "" {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		authSegs := strings.SplitN(authCode, " ", 2)
		if len(authSegs) != 2 {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		tokenStr := authSegs[1]
		var uc web.UserClaims
		token, err := jwt.ParseWithClaims(tokenStr, &uc, func(token *jwt.Token) (interface{}, error) {
			return web.JWTKey, nil
		})
		if err != nil {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		if !token.Valid {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		if uc.UserAgent != ctx.GetHeader("User-Agent") {
			//后续监控告警需要在这里埋点
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		expireTime := uc.ExpiresAt
		if expireTime.Sub(time.Now()) < time.Second*50 {
			uc.ExpiresAt = jwt.NewNumericDate(time.Now().Add(time.Second * 50))
			token := jwt.NewWithClaims(jwt.SigningMethodHS512, uc)
			tokenStr, err := token.SignedString(web.JWTKey)
			if err != nil {
				log.Println(err)
			} else {
				ctx.Header("x-jwt-token", tokenStr)
			}
		}
		ctx.Set("user", uc)
	}
}
