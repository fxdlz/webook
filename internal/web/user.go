package web

import (
	regexp "github.com/dlclark/regexp2"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
	"unicode/utf8"
	"webook/internal/domain"
	"webook/internal/service"
)

const (
	emailRegexPattern    = "^\\w+([-+.]\\w+)*@\\w+([-.]\\w+)*\\.\\w+([-.]\\w+)*$"
	passwordRegexPattern = `^(?=.*[A-Za-z])(?=.*\d)(?=.*[$@$!%*#?&])[A-Za-z\d$@$!%*#?&]{8,}$`
)

type UserHandler struct {
	svc         *service.UserService
	emailExp    *regexp.Regexp
	passwordExp *regexp.Regexp
}

func NewUserHandler(svc *service.UserService) *UserHandler {
	emailExp := regexp.MustCompile(emailRegexPattern, regexp.None)
	passwordExp := regexp.MustCompile(passwordRegexPattern, regexp.None)

	return &UserHandler{
		svc:         svc,
		emailExp:    emailExp,
		passwordExp: passwordExp,
	}
}

func (h *UserHandler) RegisterRoutes(server *gin.Engine) {
	ug := server.Group("/user")
	ug.POST("/signup", h.SignUp)
	ug.POST("/login", h.Login)
	ug.POST("/edit", h.Edit)
	ug.GET("/profile", h.Profile)
}

func (h *UserHandler) SignUp(ctx *gin.Context) {
	type SignUpReq struct {
		Email           string `json:"email"`
		Password        string `json:"password"`
		ConfirmPassword string `json:"confirmPassword"`
	}
	req := SignUpReq{}
	if err := ctx.Bind(&req); err != nil {
		return
	}
	ok, err := h.emailExp.MatchString(req.Email)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
	}
	if !ok {
		ctx.String(http.StatusOK, "邮箱不正确")
	}
	if req.Password != req.ConfirmPassword {
		ctx.String(http.StatusOK, "两次输入的密码不一致")
	}
	ok, err = h.emailExp.MatchString(req.Password)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
	}
	if !ok {
		ctx.String(http.StatusOK, "密码必须大于8位，包含数字、特殊字符")
	}
	err = h.svc.Signup(ctx.Request.Context(), domain.User{
		Email:    req.Email,
		Password: req.Password,
	})
	switch err {
	case nil:
		ctx.String(http.StatusOK, "注册成功")
	case service.ErrDuplicateEmail:
		ctx.String(http.StatusOK, "邮箱重复请换一个")
	default:
		ctx.String(http.StatusOK, "系统错误")
	}
	return

}

func (h *UserHandler) Login(ctx *gin.Context) {
	type Req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	req := Req{}
	if err := ctx.Bind(&req); err != nil {
		return
	}
	u, err := h.svc.Login(ctx.Request.Context(), req.Email, req.Password)
	switch err {
	case nil:
		sess := sessions.Default(ctx)
		sess.Set("userId", u.Id)
		sess.Options(sessions.Options{MaxAge: 900})
		err = sess.Save()
		if err != nil {
			ctx.String(http.StatusOK, "系统错误")
			return
		}
		ctx.String(http.StatusOK, "登录成功")
	case service.ErrInvalidUserOrPassword:
		ctx.String(http.StatusOK, "用户名或密码不正确")
	default:
		ctx.String(http.StatusOK, "系统错误")
	}
	return
}

func (h *UserHandler) Edit(ctx *gin.Context) {
	type EditReq struct {
		Nickname string `json:"nickname"`
		Birthday string `json:"birthday"`
		Profile  string `json:"profile"`
	}
	req := EditReq{}
	if err := ctx.Bind(&req); err != nil {
		return
	}
	if utf8.RuneCountInString(req.Nickname) > 10 {
		ctx.String(http.StatusOK, "昵称长度不能超过10")
		return
	}
	if utf8.RuneCountInString(req.Profile) > 300 {
		ctx.String(http.StatusOK, "个人简介不能超过300")
		return
	}
	layout := "2006-01-02"
	if _, err := time.Parse(layout, req.Birthday); err != nil {
		ctx.String(http.StatusOK, "日期格式不符合规范")
		return
	}
	sess := sessions.Default(ctx)
	userId := sess.Get("userId")
	v, ok := userId.(int64)
	if !ok {
		ctx.String(http.StatusOK, "系统异常")
		return
	}
	u, err := h.svc.Edit(ctx.Request.Context(), domain.User{
		Id:       v,
		Nickname: req.Nickname,
		Birthday: req.Birthday,
		Profile:  req.Profile,
	})
	switch err {
	case nil:
		ctx.JSON(http.StatusOK, u)
	default:
		ctx.String(http.StatusOK, "系统错误")
	}
	return
}

func (h *UserHandler) Profile(ctx *gin.Context) {
	sess := sessions.Default(ctx)
	userId := sess.Get("userId")
	v, ok := userId.(int64)
	if !ok {
		ctx.String(http.StatusOK, "系统异常")
		return
	}
	u, err := h.svc.Profile(ctx.Request.Context(), v)
	switch err {
	case nil:
		ctx.JSON(http.StatusOK, u)
	default:
		ctx.String(http.StatusOK, "系统错误")
	}
	return
}
