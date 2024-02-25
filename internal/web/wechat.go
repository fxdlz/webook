package web

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	uuid "github.com/lithammer/shortuuid/v4"
	"net/http"
	"webook/internal/service"
	"webook/internal/service/oauth2/wechat"
	ijwt "webook/internal/web/jwt"
)

type OAuth2WechatHandler struct {
	key             []byte
	stateCookieName string
	ijwt.Handler
	svc     wechat.Service
	userSvc service.UserService
}

func NewOAuth2WechatHandler(svc wechat.Service, userSvc service.UserService, handler ijwt.Handler) *OAuth2WechatHandler {
	return &OAuth2WechatHandler{
		svc:             svc,
		userSvc:         userSvc,
		Handler:         handler,
		stateCookieName: "jwt-state",
		key:             []byte("tD1vD9qI5bF9fX8fH5nJ6yH4FF2dD6uD"),
	}
}

func (h *OAuth2WechatHandler) RegisterRoutes(server *gin.Engine) {
	g := server.Group("/oauth2/wechat")
	g.GET("/authurl", h.OAuth2URL)
	g.Any("/callback", h.Callback)
}

func (h *OAuth2WechatHandler) OAuth2URL(ctx *gin.Context) {
	state := uuid.New()
	url, err := h.svc.AuthURL(ctx, state)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Msg:  "构造跳转URL失败",
			Code: 5,
		})
		return
	}
	err = h.setStateCookie(ctx, state)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Msg:  "系统异常",
			Code: 5,
		})
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Data: url,
	})
}

func (h *OAuth2WechatHandler) Callback(ctx *gin.Context) {
	err := h.verifyState(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Msg:  "非法请求",
			Code: 4,
		})
		return
	}
	code := ctx.Query("code")
	wechatInfo, err := h.svc.VerifyCode(ctx, code)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Msg:  "授权码有误",
			Code: 4,
		})
		return
	}
	u, err := h.userSvc.FindOrCreateByWechat(ctx, wechatInfo)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Msg:  "系统错误",
			Code: 5,
		})
		return
	}
	err = h.SetLoginToken(ctx, u.Id)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Msg:  "系统错误",
			Code: 5,
		})
	}
	ctx.JSON(http.StatusOK, Result{
		Msg: "OK",
	})
	return
}

func (h *OAuth2WechatHandler) setStateCookie(ctx *gin.Context, state string) error {
	claims := StateClaims{
		State: state,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	tokenStr, err := token.SignedString(h.key)
	if err != nil {
		return err
	}
	ctx.SetCookie(h.stateCookieName, tokenStr, 600,
		"/oauth2/wechat/callback", "", false, true)
	return nil
}

func (h *OAuth2WechatHandler) verifyState(ctx *gin.Context) error {
	state := ctx.Query("state")
	ck, err := ctx.Cookie(h.stateCookieName)
	if err != nil {
		return fmt.Errorf("无法获得Cookie, %w", err)
	}
	claims := StateClaims{}
	_, err = jwt.ParseWithClaims(ck, &claims, func(token *jwt.Token) (interface{}, error) {
		return h.key, nil
	})
	if err != nil {
		return fmt.Errorf("解析token失败,%w", err)
	}
	if state != claims.State {
		return fmt.Errorf("state 不匹配")
	}
	return nil
}

type StateClaims struct {
	jwt.RegisteredClaims
	State string
}
