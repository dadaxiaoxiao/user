package jwt

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// Handler
// Token 的相关操作·
type Handler interface {
	SetJWTToken(ctx *gin.Context, uid int64, ssid string) error
	SetLoginToken(ctx *gin.Context, uid int64) error
	ExtractToken(ctx *gin.Context) string
	CheckSession(ctx *gin.Context, ssid string) error
	ClearToken(ctx *gin.Context) error
}

type UserClaims struct {
	jwt.RegisteredClaims // 使用了组合
	Uid                  int64
	UserAgent            string
	Ssid                 string
}

type RefreshClaims struct {
	jwt.RegisteredClaims // 使用了组合
	Uid                  int64
	Ssid                 string
}
