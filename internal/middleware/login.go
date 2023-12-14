package middleware

import (
	"encoding/gob"
	"fmt"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

type LoginMiddlewareBuilder struct {
}

func (m *LoginMiddlewareBuilder) CheckLogin() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		//注册一下这个类型
		gob.Register(time.Now())
		path := ctx.Request.URL.Path
		if path == "/users/signup" || path == "/users/login" {
			return
		}
		return
		sess := sessions.Default(ctx)
		userId := sess.Get("userId")
		if userId == nil {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		const updateTimeKey = "update_time"
		now := time.Now()
		val := sess.Get(updateTimeKey)
		lastUpdateTime, ok := val.(time.Time)
		if val != nil || !ok || now.Sub(lastUpdateTime) > time.Second*30 {
			sess.Set(updateTimeKey, now)
			sess.Set("userId", userId)
			err := sess.Save()
			if err != nil {
				fmt.Println(err)
			}
		}
	}
}
