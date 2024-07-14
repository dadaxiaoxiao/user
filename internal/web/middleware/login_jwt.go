package middleware

import (
	"encoding/gob"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"time"
	myjwt "github.com/dadaxiaoxiao/user/internal/web/jwt"
)

type LoginJWTMiddlewareBuilder struct {
	paths []string
	myjwt.Handler
}

// IgnorePaths 忽略路径
func (l *LoginJWTMiddlewareBuilder) IgnorePaths(path string) *LoginJWTMiddlewareBuilder {
	l.paths = append(l.paths, path)
	return l
}

// NewLoginJWTMiddlewareBuilder 返回实例
func NewLoginJWTMiddlewareBuilder(wtHdl myjwt.Handler) *LoginJWTMiddlewareBuilder {
	return &LoginJWTMiddlewareBuilder{
		Handler: wtHdl,
	}
}

// Build 生成中间件
func (l *LoginJWTMiddlewareBuilder) Build() gin.HandlerFunc {
	// 用 Go 的方式编码解码
	gob.Register(time.Now())
	return func(ctx *gin.Context) {
		for _, path := range l.paths {
			if ctx.Request.URL.Path == path {
				return // 不需要校验 cookie + session
			}
		}

		tokenStr := l.Handler.ExtractToken(ctx)
		if tokenStr == "" {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		claims := myjwt.UserClaims{}
		token, err := jwt.ParseWithClaims(tokenStr, &claims, func(token *jwt.Token) (interface{}, error) {
			return myjwt.AccessTokenKey, nil
		})
		if err != nil {
			// 没登录
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		// err 为 nil，token 不为 nil
		if token == nil || !token.Valid || claims.Uid == 0 {
			// 没登录
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		if claims.UserAgent != ctx.Request.UserAgent() {
			// 安全问题
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		err = l.Handler.CheckSession(ctx, claims.Ssid)
		if err != nil {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		ctx.Set("user", claims)
	}
}
