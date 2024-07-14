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

var (
	AccessTokenKey  = []byte("95osj3fUD7fo0mlYdDbncXz4VD2igvf0")
	RefreshTokenKey = []byte("95osj3fUD7fo0mlYdDbncXz4VD2igvf0")
)

type RedisJWTHandler struct {
	cmd          redis.Cmdable
	rcExpiration time.Duration
}

func NewRedisJWTHandler(cmd redis.Cmdable) Handler {
	return &RedisJWTHandler{
		cmd:          cmd,
		rcExpiration: time.Hour * 24 * 7,
	}
}

func (r *RedisJWTHandler) SetLoginToken(ctx *gin.Context, uid int64) error {
	ssid := uuid.New().String()
	// 设置 jwt token
	err := r.SetJWTToken(ctx, uid, ssid)
	if err != nil {
		return err
	}
	// 设置 refreshtoken
	err = r.setRefreshToken(ctx, uid, ssid)
	return err
}

func (r *RedisJWTHandler) SetJWTToken(ctx *gin.Context, uid int64, ssid string) error {
	claims := UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
		Uid:       uid,
		UserAgent: ctx.Request.UserAgent(),
		Ssid:      ssid,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims) // token 携带用户信息

	tokenStr, err := token.SignedString(AccessTokenKey)
	if err != nil {
		return err
	}
	ctx.Header("x-jwt-token", tokenStr)
	return nil
}

// setRefreshToken 设置refresh
func (r *RedisJWTHandler) setRefreshToken(ctx *gin.Context, uid int64, ssid string) error {
	claims := RefreshClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			// 有效期7 天
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(r.rcExpiration)),
		},
		Uid:  uid,
		Ssid: ssid,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims) // token 携带用户信息

	tokenStr, err := token.SignedString(RefreshTokenKey)
	if err != nil {
		return err
	}
	ctx.Header("x-refresh-token", tokenStr)
	return nil
}

// ExtractToken 提取jwt token
func (r *RedisJWTHandler) ExtractToken(ctx *gin.Context) string {
	tokenHeader := ctx.GetHeader("Authorization") // 	Bearer token
	segs := strings.Split(tokenHeader, " ")       //根据 “ ” 切割得的数组
	if len(segs) != 2 {

		return ""
	}
	return segs[1] // 得到token 值
}

func (r *RedisJWTHandler) CheckSession(ctx *gin.Context, ssid string) error {
	cnt, err := r.cmd.Exists(ctx, fmt.Sprintf("users:ssid:%s", ssid)).Result()
	switch err {
	case redis.Nil: // key 不存在
		return nil
	case nil:
		if cnt == 0 {
			return nil
		}
		// 存在key
		return errors.New("session 已经无效了")
	default:
		return err
	}
	return nil
}

func (r *RedisJWTHandler) ClearToken(ctx *gin.Context) error {
	// 前端用户 会把两个token 更新
	// 这样 登录校验里面，走不到查询redis
	ctx.Header("x-jwt-token", "")
	ctx.Header("x-refresh-token", "")
	// 获取 jwt token 中间件解析的  UserClaims

	c, _ := ctx.Get("user")
	uc, ok := c.(*UserClaims)
	if !ok {
		return errors.New("解析UserClaims 错误")
	}

	return r.cmd.Set(ctx, fmt.Sprintf("users:ssid:%s", uc.Ssid), "", r.rcExpiration).Err()
}
