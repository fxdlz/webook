package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"strings"
	ijwt "webook/internal/web/jwt"
)

type LoginJWTMiddlewareBuilder struct {
	ijwt.Handler
}

func NewLoginJWTMiddlewareBuilder(handler ijwt.Handler) *LoginJWTMiddlewareBuilder {
	return &LoginJWTMiddlewareBuilder{
		Handler: handler,
	}
}

func (m *LoginJWTMiddlewareBuilder) CheckLogin() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		path := ctx.Request.URL.Path
		if path == "/users/signup" ||
			path == "/users/login" ||
			path == "/users/login_sms/code/send" ||
			path == "/users/login_sms" ||
			path == "/oauth2/wechat/authurl" ||
			path == "/oauth2/wechat/callback" ||
			path == "/oauth2/wechat/refresh_token" ||
			strings.HasPrefix(path, "/articles/pub/like-top/") {
			return
		}
		tokenStr := m.ExtractToken(ctx)
		var uc ijwt.UserClaims
		token, err := jwt.ParseWithClaims(tokenStr, &uc, func(token *jwt.Token) (interface{}, error) {
			return ijwt.JWTKey, nil
		})
		if err != nil {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		if !token.Valid {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		err = m.CheckSession(ctx, uc.Ssid)

		if err != nil {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		//if uc.UserAgent != ctx.GetHeader("User-Agent") {
		//	//后续监控告警需要在这里埋点
		//	ctx.AbortWithStatus(http.StatusUnauthorized)
		//	return
		//}
		//expireTime := uc.ExpiresAt
		//if expireTime.Sub(time.Now()) < time.Second*50 {
		//	uc.ExpiresAt = jwt.NewNumericDate(time.Now().Add(time.Second * 50))
		//	token := jwt.NewWithClaims(jwt.SigningMethodHS512, uc)
		//	tokenStr, err := token.SignedString(web.JWTKey)
		//	if err != nil {
		//		log.Println(err)
		//	} else {
		//		ctx.Header("x-jwt-token", tokenStr)
		//	}
		//}
		ctx.Set("user", uc)
	}
}
