package web

import (
	"errors"
	"fmt"
	"github.com/dadaxiaoxiao/user/internal/service"
	"github.com/dadaxiaoxiao/user/internal/service/oauth2/wechat"
	myjwt "github.com/dadaxiaoxiao/user/internal/web/jwt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	uuid "github.com/lithammer/shortuuid/v4"
	"net/http"
	"time"
)

type OAuth2WechatHandler struct {
	svc     wechat.Service
	usersvc service.UserService
	myjwt.Handler
	stateKey []byte
	cfg      WechatHandlerConfig
}

type WechatHandlerConfig struct {
	Secure bool
}

func NewOAuth2WechatHandler(svc wechat.Service, usersvc service.UserService, cfg WechatHandlerConfig, wtHdl myjwt.Handler) *OAuth2WechatHandler {
	return &OAuth2WechatHandler{
		svc:      svc,
		usersvc:  usersvc,
		Handler:  wtHdl,
		stateKey: []byte("95osj3fUD7foxmlYdDbncXz4VD2igvf1"),
		cfg:      cfg,
	}
}

func (h *OAuth2WechatHandler) RegisterRoutes(s *gin.Engine) {
	g := s.Group("/oauth2/wechat")
	g.GET("/authurl", h.OAuth2URL)
	g.GET("/callback", h.Callback)
}

func (h *OAuth2WechatHandler) OAuth2URL(ctx *gin.Context) {
	state := uuid.New()
	url, err := h.svc.AuthURL(ctx, state)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "构造扫码登录URL失败",
		})
	}
	//// 设置 state
	//if err = h.setStateCookie(ctx, state); err != nil {
	//	ctx.JSON(http.StatusOK, Result{
	//		Code: 5,
	//		Msg:  "系统异常",
	//	})
	//}

	ctx.JSONP(http.StatusOK, Result{
		Data: url,
	})
}

func (h *OAuth2WechatHandler) Callback(ctx *gin.Context) {
	code := ctx.Query("code")
	//err := h.verifyState(ctx)
	//if err != nil {
	//	ctx.JSONP(http.StatusOK, Result{
	//		Code: 5,
	//		Msg:  "登录失败",
	//	})
	//}

	info, err := h.svc.VerifyCode(ctx, code)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
	}

	// 查找或新创建用户
	user, err := h.usersvc.FindOrCreateByWechat(ctx, info)

	if err != nil {
		ctx.JSONP(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}

	// 设置token
	if err = h.SetLoginToken(ctx, user.Id); err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}

	//登录成功
	ctx.JSONP(http.StatusOK, Result{
		Msg: "登录成功",
	})
}

func (h *OAuth2WechatHandler) setStateCookie(ctx *gin.Context, state string) error {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, StateClaims{
		State: state,
		RegisteredClaims: jwt.RegisteredClaims{
			// 过期时间，你预期中一个用户完成登录的时间
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 10)),
		},
	})
	tokenStr, err := token.SignedString(h.stateKey)
	if err != nil {
		return err
	}
	ctx.SetCookie("jwt-state", tokenStr,
		600, "/oauth2/wechat/callback",
		"", h.cfg.Secure, true)
	return nil
}

// verifyState 校验 state
func (h *OAuth2WechatHandler) verifyState(ctx *gin.Context) error {
	// 从 “/oauth2/wechat/callback” 获取 state 参数
	state := ctx.Query("state")
	// 检查cookie
	ck, err := ctx.Cookie("jwt-state")
	if err != nil {
		return fmt.Errorf("找不到 state 的cookie,%w", err)
	}
	var claims StateClaims
	token, err := jwt.ParseWithClaims(ck, &claims, func(token *jwt.Token) (interface{}, error) {
		return h.stateKey, nil
	})
	if err != nil || !token.Valid {
		return fmt.Errorf("token 已经过期了, %w", err)
	}
	if claims.State != state {
		return errors.New("state 不相等")
	}
	return nil
}

type StateClaims struct {
	State string
	jwt.RegisteredClaims
}
