package service

import (
	"context"
	"errors"
	"github.com/dadaxiaoxiao/go-pkg/accesslog"
	"github.com/dadaxiaoxiao/user/internal/domain"
	"github.com/dadaxiaoxiao/user/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserDuplicateEmail    = repository.ErrUserDuplicateEmail
	ErrUserNotFound          = repository.ErrUserNotFound
	ErrInvalidUserOrPassword = errors.New("账号/邮箱或密码不对")
)

type UserService interface {
	Signup(ctx context.Context, user domain.User) error
	FindOrCreate(ctx context.Context, phone string) (user domain.User, err error)
	FindOrCreateByWechat(ctx context.Context, info domain.WechatInfo) (user domain.User, err error)
	Login(ctx context.Context, email, password string) (domain.User, error)
	UpdateNonSensitiveInfo(ctx context.Context, user domain.User) error
	Profile(ctx context.Context, id int64) (domain.User, error)
}

type userService struct {
	repo repository.UserRepository
	log  accesslog.Logger
}

// NewUserService 实现UserService 接口的实例
func NewUserService(repo repository.UserRepository, log accesslog.Logger) UserService {
	return &userService{
		repo: repo,
		log:  log,
	}
}

// Signup 业务层注册
func (svc *userService) Signup(ctx context.Context, user domain.User) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(hash) // 字节转字符串
	return svc.repo.Create(ctx, user)
}

func (svc *userService) FindOrCreate(ctx context.Context, phone string) (user domain.User, err error) {
	// 1.先查询，如果存在，直接返回
	u, err := svc.repo.FindByPhone(ctx, phone)
	if err != repository.ErrUserNotFound {
		return u, err
	}

	// 2. 注册一个用户
	u = domain.User{
		Phone: phone,
	}

	err = svc.repo.Create(ctx, u)
	if err != nil && err != repository.ErrUserDuplicateEmail {
		return u, err
	}

	// 3.查询数据库
	// 会遇到主从延迟问题
	return svc.repo.FindByPhone(ctx, phone)
}

// FindOrCreateByWechat 根据微信查询新建用户
func (svc *userService) FindOrCreateByWechat(ctx context.Context, info domain.WechatInfo) (user domain.User, err error) {
	u, err := svc.repo.FindByWechat(ctx, info.OpenId)
	if err != repository.ErrUserNotFound {
		return u, err
	}

	// 直接使用包变量
	//zap.L().Info("微信用户未注册，注册新用户",
	//	zap.Any("wechat_info", info))

	// 使用依赖注入的logger
	svc.log.Info("微信用户未注册，注册新用户",
		accesslog.Any("wechat_info", info))

	// 2. 注册一个用户
	u = domain.User{
		WechatInfo: info,
	}
	// 注册一个用户
	err = svc.repo.Create(ctx, u)
	if err != nil {
		return u, err
	}
	// 会遇到主从延迟问题
	return svc.repo.FindByWechat(ctx, info.OpenId)
}

// Login 用户登录，返回domain.User ,error
func (svc *userService) Login(ctx context.Context, email, password string) (domain.User, error) {
	// 查询email 对应的 用户信息
	u, err := svc.repo.FindByEmail(ctx, email)

	if err == repository.ErrUserNotFound {
		return domain.User{}, ErrInvalidUserOrPassword
	}
	if err != nil {
		return domain.User{}, err
	}

	// 密码比较
	err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))

	if err != nil {
		return domain.User{}, ErrInvalidUserOrPassword // 密码不对
	}
	return u, nil
}

// UpdateNonSensitiveInfo 修改用户信息
func (svc *userService) UpdateNonSensitiveInfo(ctx context.Context, user domain.User) error {
	u, err := svc.repo.FindById(ctx, user.Id)
	if err != nil {
		return ErrUserNotFound
	}
	if u.Id == 0 {
		return ErrUserNotFound
	}

	// 这种是复杂写法，依赖于 repository 中更新会忽略 0 值
	// 这个转换的意义在于，你在 service 层面上维护住了什么是敏感字段这个语义
	user.Email = ""
	user.Phone = ""
	user.Password = ""
	user.WechatInfo = domain.WechatInfo{}
	return svc.repo.Update(ctx, user)
}

// Profile 查询个人信息
func (svc *userService) Profile(ctx context.Context, id int64) (domain.User, error) {
	u, err := svc.repo.FindById(ctx, id)
	if err != nil {
		return domain.User{}, ErrUserNotFound
	}
	return u, nil
}
