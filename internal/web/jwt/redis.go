package jwt

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"strings"
	"time"
)

type RedisJWTHandler struct {
	client        redis.Cmdable
	signingMethod jwt.SigningMethod
	rcExpiration  time.Duration
}

var RefreshTokenKey = []byte("tD1vD9qI5bF9fX8fH5nJ6yH4FF2dD6uX")
var JWTKey = []byte("tD1vD9qI5bF9fX8fH5nJ6yH4FF2dD6uX")

type RefreshClaims struct {
	jwt.RegisteredClaims
	Uid  int64
	Ssid string
}

type UserClaims struct {
	jwt.RegisteredClaims
	Uid       int64
	UserAgent string
	Ssid      string
}

func NewRedisJWTHandler(client redis.Cmdable) Handler {
	return &RedisJWTHandler{
		signingMethod: jwt.SigningMethodHS512,
		client:        client,
		rcExpiration:  time.Hour * 24 * 7,
	}
}

func (h *RedisJWTHandler) SetJWTToken(ctx *gin.Context, uid int64, ssid string) error {
	uc := UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Second * 90)),
		},
		Uid:       uid,
		UserAgent: ctx.GetHeader("User-Agent"),
		Ssid:      ssid,
	}
	token := jwt.NewWithClaims(h.signingMethod, uc)
	tokenStr, err := token.SignedString(JWTKey)
	if err != nil {
		return err
	}
	ctx.Header("x-jwt-token", tokenStr)
	return nil
}

func (h *RedisJWTHandler) SetRefreshToken(ctx *gin.Context, uid int64, ssid string) error {
	rc := RefreshClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(h.rcExpiration)),
		},
		Uid:  uid,
		Ssid: ssid,
	}
	token := jwt.NewWithClaims(h.signingMethod, rc)
	tokenStr, err := token.SignedString(RefreshTokenKey)
	if err != nil {
		return err
	}
	ctx.Header("x-refresh-token", tokenStr)
	return nil
}

// ExtractToken 根据约定token在Authorization头部
func (h *RedisJWTHandler) ExtractToken(ctx *gin.Context) string {
	authCode := ctx.GetHeader("Authorization")
	//没登录，没有token
	if authCode == "" {
		return authCode
	}
	authSegs := strings.SplitN(authCode, " ", 2)
	if len(authSegs) != 2 {
		return ""
	}
	return authSegs[1]
}

func (h *RedisJWTHandler) SetLoginToken(ctx *gin.Context, uid int64) error {
	ssid := uuid.New().String()
	err := h.SetRefreshToken(ctx, uid, ssid)
	if err != nil {
		return err
	}
	return h.SetJWTToken(ctx, uid, ssid)
}

func (h *RedisJWTHandler) CheckSession(ctx *gin.Context, ssid string) error {
	cnt, err := h.client.Exists(ctx, fmt.Sprintf("users:ssid:%s", ssid)).Result()
	if err != nil {
		return err
	}
	if cnt > 0 {
		return errors.New("用户已退出")
	}
	return nil
}

func (h *RedisJWTHandler) ClearToken(ctx *gin.Context) error {
	ctx.Header("x-refresh-token", "")
	ctx.Header("x-jwt-token", "")
	uc := ctx.MustGet("user").(UserClaims)
	return h.client.Set(ctx, fmt.Sprintf("users:ssid:%s", uc.Ssid), "", h.rcExpiration).Err()
}
