package web

import (
	"github.com/dadaxiaoxiao/go-pkg/accesslog"
	"github.com/dadaxiaoxiao/user/internal/domain"
	"github.com/dadaxiaoxiao/user/internal/errs"
	"github.com/dadaxiaoxiao/user/internal/service"
	myjwt "github.com/dadaxiaoxiao/user/internal/web/jwt"
	regexp "github.com/dlclark/regexp2"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.opentelemetry.io/otel/trace"
	"net/http"
	"time"
	"unicode/utf8"
)

const (
	emailRegexPattern    = "^\\w+([-+.]\\w+)*@\\w+([-.]\\w+)*\\.\\w+([-.]\\w+)*$"
	passwordRegexPattern = `^(?=.*[A-Za-z])(?=.*\d)(?=.*[$@$!%*#?&])[A-Za-z\d$@$!%*#?&]{8,72}$`
	birthdayRegexPattern = `^(?:(?:1[89]|20)\d\d)-(?:0[1-9]|1[0-2])-(?:0[1-9]|[12]\d|3[01])$`
	phoneRegexPattern    = `^1[3456789]\d{9}$`
	biz                  = "login"
)

// UserHandler  定义跟用户有关的路由
type UserHandler struct {
	userSvc          service.UserService
	codeSvc          service.CodeService
	emailRegexExp    *regexp.Regexp
	passwordRegexExp *regexp.Regexp
	birthdayRegexExp *regexp.Regexp
	phoneRegexExp    *regexp.Regexp
	log              accesslog.Logger
	myjwt.Handler
}

// NewUserHandler 返回 UserHandler 类的指针
func NewUserHandler(svc service.UserService, codeSvc service.CodeService, wtHdl myjwt.Handler, log accesslog.Logger) *UserHandler {
	return &UserHandler{
		userSvc:          svc,
		codeSvc:          codeSvc,
		emailRegexExp:    regexp.MustCompile(emailRegexPattern, regexp.None),
		passwordRegexExp: regexp.MustCompile(passwordRegexPattern, regexp.None),
		birthdayRegexExp: regexp.MustCompile(birthdayRegexPattern, regexp.None),
		phoneRegexExp:    regexp.MustCompile(phoneRegexPattern, regexp.None),
		Handler:          wtHdl,
		log:              log,
	}
}

// RegisterRoutes 注册路由
func (u *UserHandler) RegisterRoutes(server *gin.Engine) {
	// 路由分组
	ug := server.Group("/users")

	ug.POST("/signup", u.Signup)
	ug.POST("/edit", u.Edit)
	ug.POST("/login", u.LoginJWT)
	ug.GET("/profile", u.Profile)
	ug.POST("/logout", u.Logout)
	ug.POST("/login_sms/code/send", u.SendSMSLoginCode)
	ug.POST("/login_sms", u.LoginSMS)
	ug.POST("/refresh_token", u.RefreshToken)

}

// Signup 注册
func (u *UserHandler) Signup(ctx *gin.Context) {
	// 方法内部类
	type SignUpReq struct {
		Email           string `json:"email"`
		ConfirmPassword string `json:"confirmPassword"`
		Password        string `json:"password"`
	}

	var req SignUpReq
	// Bind 方法会根据Content-Type 来解析 到结构体里面 所以传地址
	// 解析错了，会返回状态码 4xx
	if err := ctx.Bind(&req); err != nil {
		return
	}

	// 判断邮箱正则表达式
	isEmail, err := u.emailRegexExp.MatchString(req.Email)
	// 判断系统错误
	if err != nil {
		ctx.JSONP(http.StatusOK, Result{
			Code: 4,
			Msg:  "系统错误",
		})
		return
	}
	// 判断是否为邮箱
	if !isEmail {
		ctx.JSONP(http.StatusOK, Result{
			Code: 4,
			Msg:  "邮箱不正确",
		})
		return
	}

	// 判断两次输入的密码
	if req.Password != req.ConfirmPassword {
		ctx.JSONP(http.StatusOK, Result{
			Code: 4,
			Msg:  "两次输入的密码不相同",
		})
		return
	}

	// 判断密码是否符合规则
	isPassword, err := u.passwordRegexExp.MatchString(req.Password)
	if err != nil {
		ctx.JSONP(http.StatusOK, Result{
			Code: 4,
			Msg:  "系统错误",
		})
		return
	}
	if !isPassword {
		ctx.JSONP(http.StatusOK, Result{
			Code: 4,
			Msg:  "密码必须包含数字、特殊字符，并且长度不能小于 8 位",
		})
		return
	}
	err = u.userSvc.Signup(ctx.Request.Context(), domain.User{
		Email:    req.Email,
		Password: req.Password,
	})
	if err == service.ErrUserDuplicateEmail {
		span := trace.SpanFromContext(ctx.Request.Context())
		span.AddEvent("邮件冲突")
		ctx.JSON(http.StatusOK, Result{Code: 4, Msg: "邮箱冲突"})
		return
	}

	if err != nil {
		ctx.JSON(http.StatusOK, Result{Code: 4, Msg: "系统异常"})
		return
	}
	ctx.JSON(http.StatusOK, Result{Msg: "注册成功"})
}


// LoginJWT 登录 得到jwt token
func (u *UserHandler) LoginJWT(ctx *gin.Context) {
	type LoginReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var req LoginReq
	if err := ctx.Bind(&req); err != nil {
		return
	}

	user, err := u.userSvc.Login(ctx, req.Email, req.Password)
	if err == service.ErrInvalidUserOrPassword {
		ctx.JSONP(http.StatusOK, Result{
			Code: errs.UserInvalidOrPassword,
			Msg:  "用户名或密码不对",
		})
		return
	}
	if err != nil {
		ctx.JSONP(http.StatusOK, Result{
			Code: 4,
			Msg:  "系统异常",
		})
		return
	}

	err = u.SetLoginToken(ctx, user.Id)
	if err != nil {
		ctx.JSONP(http.StatusOK, Result{
			Code: 4,
			Msg:  "系统异常",
		})
	}

	ctx.JSONP(http.StatusOK, Result{
		Msg: "登录成功",
	})
	return
}

// RefreshToken 刷新token
func (u *UserHandler) RefreshToken(ctx *gin.Context) {
	// 要求前端 请求刷新token 接口时候，一样通过 Authorization 传值
	refreshToken := u.Handler.ExtractToken(ctx)
	var claims myjwt.RefreshClaims
	token, err := jwt.ParseWithClaims(refreshToken, &claims, func(token *jwt.Token) (interface{}, error) {
		return myjwt.RefreshTokenKey, nil
	})
	if err != nil || !token.Valid {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	//检查ssid
	err = u.Handler.CheckSession(ctx, claims.Ssid)
	if err != nil {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	// 重新生成token
	err = u.Handler.SetJWTToken(ctx, claims.Uid, claims.Ssid)
	if err != nil {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	ctx.JSONP(http.StatusOK, Result{
		Msg: "刷新成功",
	})
}

// Logout 登出
func (u *UserHandler) Logout(ctx *gin.Context) {
	err := u.Handler.ClearToken(ctx)
	if err != nil {
		ctx.JSONP(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	ctx.JSONP(http.StatusOK, Result{
		Msg: "退出登录成功",
	})
}

// Edit 编辑
func (u *UserHandler) Edit(ctx *gin.Context) {
	type EditReq struct {
		Nickname string `json:"nickname"`
		Birthday string `json:"birthday"`
		AboutMe  string `json:"aboutMe"`
	}
	var req EditReq
	if err := ctx.Bind(&req); err != nil {
		ctx.JSONP(http.StatusOK, Result{
			Code: 4,
			Msg:  "参数异常",
		})
		return
	}

	if req.Nickname == "" {
		ctx.JSON(http.StatusOK, Result{Code: 4, Msg: "昵称不能为空"})
		return
	}

	// 判断生日格式
	isBirthday, err := u.birthdayRegexExp.MatchString(req.Birthday)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{Code: 4, Msg: "系统错误"})
		return
	}
	if !isBirthday {
		ctx.JSON(http.StatusOK, Result{Code: 4, Msg: "日期格式不对，请输入 YYYY-MM-DD 的日期格式"})
		return
	}
	birthday, _ := time.Parse(time.DateOnly, req.Birthday)

	// 判断简介长度
	if utf8.RuneCountInString(req.AboutMe) > 1024 {
		ctx.JSON(http.StatusOK, Result{Code: 4, Msg: "个人简介过长"})
		return
	}

	uc := ctx.MustGet("user").(myjwt.UserClaims)

	err = u.userSvc.UpdateNonSensitiveInfo(ctx, domain.User{
		Id:       uc.Uid,
		Nickname: req.Nickname,
		Birthday: birthday,
		AboutMe:  req.AboutMe,
	})

	if err != nil {
		ctx.JSON(http.StatusOK, Result{Code: 4, Msg: "系统错误"})
		return
	}

	ctx.JSON(http.StatusOK, Result{Msg: "更新成功"})
}

// Profile 查询
func (u *UserHandler) Profile(ctx *gin.Context) {
	// Id 通过jwt 获取
	c, _ := ctx.Get("user")
	claims, ok := c.(myjwt.UserClaims)
	if !ok {
		ctx.JSON(http.StatusOK, Result{Code: 4, Msg: "系统错误"})
	}

	user, err := u.userSvc.Profile(ctx, claims.Uid)
	if err == service.ErrUserNotFound {
		ctx.JSON(http.StatusOK, Result{Code: 4, Msg: "无法查询用户相关信息"})
		return
	}
	if err != nil {
		ctx.JSON(http.StatusOK, Result{Code: 4, Msg: "系统错误"})
		return
	}
	type rep struct {
		Email    string
		Phone    string
		Nickname string
		Birthday string
		AboutMe  string
	}

	ctx.JSONP(http.StatusOK, Result{Data: rep{
		Email:    user.Email,
		Phone:    user.Phone,
		Nickname: user.Nickname,
		Birthday: user.Birthday.Format(time.DateOnly),
		AboutMe:  user.AboutMe,
	}})
}

// SendSMSLoginCode 发送短信登录验证码
func (u *UserHandler) SendSMSLoginCode(ctx *gin.Context) {
	type Req struct {
		Phone string `json:"phone"`
	}

	var req Req
	if err := ctx.Bind(&req); err != nil {
		ctx.JSONP(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
	}

	if req.Phone == "" {
		ctx.JSONP(http.StatusOK, Result{
			Code: 4,
			Msg:  "请输入手机号",
		})
	}

	// 判断手机号格式
	isPhone, err := u.phoneRegexExp.MatchString(req.Phone)
	if err != nil {
		ctx.JSONP(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	if !isPhone {
		ctx.JSONP(http.StatusOK, Result{
			Code: 4,
			Msg:  "不是有效的手机号",
		})
		return
	}

	// 发送验证码
	err = u.codeSvc.Send(ctx.Request.Context(), biz, req.Phone)
	switch err {
	case nil:
		// 发送成功
		ctx.JSONP(http.StatusOK, Result{
			Msg: "发送成功",
		})
	case service.ErrCodeSendTooMany:
		// 发送太频繁
		ctx.JSONP(http.StatusOK, Result{
			Msg: "短信发送太频繁，请稍后再试",
		})
		u.log.Warn("短信发送太频繁", accesslog.Any("error", err))
	default:
		ctx.JSONP(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		u.log.Warn("短信发送失败", accesslog.Any("error", err))
	}
	return
}

// LoginSMS 登录短信校验
func (u *UserHandler) LoginSMS(ctx *gin.Context) {
	type Req struct {
		Phone string `json:"phone"`
		Code  string `json:"code"`
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		ctx.JSONP(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	// 判断手机号格式
	isPhone, err := u.phoneRegexExp.MatchString(req.Phone)
	if err != nil {
		ctx.JSONP(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	if !isPhone {
		ctx.JSONP(http.StatusOK, Result{
			Code: 4,
			Msg:  "不是有效的手机号",
		})
		return
	}

	// 验证手机号验证码
	ok, err := u.codeSvc.Verify(ctx, biz, req.Phone, req.Code)
	if err != nil {
		ctx.JSONP(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		u.log.Error("校验验证码出错",
			accesslog.Any("error", err),
			accesslog.String("手机号", req.Phone)) // 手机号是敏感数据，不建议打印到日志里面

		return
	}
	if !ok {
		ctx.JSONP(http.StatusOK, Result{
			Code: 4,
			Msg:  "验证码错误",
		})
		return
	}

	// 查找或新创建用户
	user, err := u.userSvc.FindOrCreate(ctx, req.Phone)
	if err != nil {
		ctx.JSONP(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}

	// 设置token
	if err = u.SetLoginToken(ctx, user.Id); err != nil {
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
