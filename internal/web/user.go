package web

import (
	regexp "github.com/dlclark/regexp2"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"time"
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
	ug := server.Group("/users")
	ug.POST("/signup", h.SignUp)
	//ug.POST("/login", h.Login)
	ug.POST("/login", h.LoginJWT)
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
		return
	}
	if !ok {
		ctx.String(http.StatusOK, "邮箱不正确")
		return
	}
	if req.Password != req.ConfirmPassword {
		ctx.String(http.StatusOK, "两次输入的密码不一致")
		return
	}
	ok, err = h.passwordExp.MatchString(req.Password)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	if !ok {
		ctx.String(http.StatusOK, "密码必须大于8位，包含数字、特殊字符")
		return
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

func (h *UserHandler) LoginJWT(ctx *gin.Context) {
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
		uc := UserClaims{
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Second * 90)),
			},
			Uid:       u.Id,
			UserAgent: ctx.GetHeader("User-Agent"),
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS512, uc)
		tokenStr, err := token.SignedString(JWTKey)
		if err != nil {
			ctx.String(http.StatusOK, "系统错误")
		}
		ctx.Header("x-jwt-token", tokenStr)
		ctx.String(http.StatusOK, "登录成功")
	case service.ErrInvalidUserOrPassword:
		ctx.String(http.StatusOK, "用户名或密码不正确")
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
		sess.Options(sessions.Options{MaxAge: 90})
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

type UserInfo struct {
	Email    string
	Nickname string
	Birthday string
	Profile  string
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
	err := h.svc.Edit(ctx.Request.Context(), domain.User{
		Id:       v,
		Nickname: req.Nickname,
		Birthday: req.Birthday,
		Profile:  req.Profile,
	})
	switch err {
	case nil:
		ctx.JSON(http.StatusOK, "修改成功")
	default:
		ctx.String(http.StatusOK, "系统错误")
	}
	return
}

func (h *UserHandler) Profile(ctx *gin.Context) {
	uc := ctx.MustGet("user").(UserClaims)
	v := uc.Uid
	//sess := sessions.Default(ctx)
	//userId := sess.Get("userId")
	//v, ok := userId.(int64)
	//if !ok {
	//	ctx.String(http.StatusOK, "系统异常")
	//	return
	//}
	u, err := h.svc.Profile(ctx.Request.Context(), v)

	switch err {
	case nil:
		ctx.JSON(http.StatusOK, UserInfo{
			Email:    u.Email,
			Nickname: u.Nickname,
			Birthday: u.Birthday,
			Profile:  u.Profile,
		})
	default:
		ctx.String(http.StatusOK, "系统错误")
	}
	return
}

type UserClaims struct {
	jwt.RegisteredClaims
	Uid       int64
	UserAgent string
}

var JWTKey = []byte("tD1vD9qI5bF9fX8fH5nJ6yH4FF2dD6uM")
