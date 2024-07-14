package repository

import (
	"context"
	"database/sql"
	"github.com/dadaxiaoxiao/user/internal/domain"
	"github.com/dadaxiaoxiao/user/internal/repository/cache"
	"github.com/dadaxiaoxiao/user/internal/repository/dao"
	"time"
)

var (
	ErrUserDuplicateEmail = dao.ErrUserDuplicateEmail
	ErrUserNotFound       = dao.ErrUserNotFound
)

//go:generate mockgen.exe -source=./user.go -package=repomocks -destination=mocks/user.mock.go UserRepository
type UserRepository interface {
	Create(ctx context.Context, user domain.User) error
	FindByEmail(ctx context.Context, email string) (domain.User, error)
	FindByPhone(ctx context.Context, phone string) (domain.User, error)
	FindById(ctx context.Context, id int64) (domain.User, error)
	Update(ctx context.Context, user domain.User) error
	FindByWechat(ctx context.Context, openID string) (domain.User, error)
}

type CachedUserRepository struct {
	dao   dao.UserDao
	cache cache.UserCache
}

// NewCachedUserRepository 使用了缓存的 UserRepository 实现
func NewCachedUserRepository(dao dao.UserDao, cache cache.UserCache) UserRepository {
	return &CachedUserRepository{
		dao:   dao,
		cache: cache,
	}
}

// Create 数据存储层新增用户
func (r *CachedUserRepository) Create(ctx context.Context, user domain.User) error {
	return r.dao.Insert(ctx, r.domainToEntity(user))
}

// FindByEmail 根据email 查询用信息
func (r *CachedUserRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	u, err := r.dao.FindByEmail(ctx, email)
	if err != nil {
		return domain.User{}, err
	}
	return r.entityToDomain(u), nil
}

// FindByPhone 根据 phone 查找用户信息
func (r *CachedUserRepository) FindByPhone(ctx context.Context, phone string) (domain.User, error) {
	u, err := r.dao.FindByPhone(ctx, phone)
	if err != nil {
		return domain.User{}, err
	}
	return r.entityToDomain(u), err
}

func (r *CachedUserRepository) FindByWechat(ctx context.Context, openID string) (domain.User, error) {
	u, err := r.dao.FindByWechat(ctx, openID)
	if err != nil {
		return domain.User{}, err
	}
	return r.entityToDomain(u), err
}

// FindById 根据id 查询用户信息
func (r *CachedUserRepository) FindById(ctx context.Context, id int64) (domain.User, error) {
	// 获取缓存
	u, err := r.cache.Get(ctx, id)
	if err == nil {
		// 必然是有数据
		return u, nil
	}

	// 没有缓存，查询数据库
	user, err := r.dao.FindById(ctx, id)
	if err != nil {
		return domain.User{}, err
	}
	u = r.entityToDomain(user)
	// 写入缓存
	//go func() {
	//	err = r.cache.Set(ctx, u)
	//	if err != nil {
	//		// 日志写入
	//	}
	//}()

	err = r.cache.Set(ctx, u)
	if err != nil {
		// 日志写入
	}

	return u, err
}

// Update 修改信息
func (r *CachedUserRepository) Update(ctx context.Context, user domain.User) error {
	err := r.dao.UpdateNonZeroFields(ctx, r.domainToEntity(user))
	if err != nil {
		return err
	}
	return r.cache.Delete(ctx, user.Id)
}

func (r *CachedUserRepository) domainToEntity(u domain.User) dao.User {
	return dao.User{
		Id: u.Id,
		Email: sql.NullString{
			String: u.Email,
			Valid:  u.Email != "",
		},
		Phone: sql.NullString{
			String: u.Phone,
			Valid:  u.Phone != "",
		},
		Password: u.Password,
		Nickname: sql.NullString{
			String: u.Nickname,
			Valid:  u.Nickname != "",
		},
		Birthday: sql.NullInt64{
			Int64: u.Birthday.UnixMilli(),
			Valid: u.Birthday.IsZero(),
		},
		AboutMe: sql.NullString{
			String: u.AboutMe,
			Valid:  u.AboutMe != "",
		},
		WechatOpenId: sql.NullString{
			String: u.WechatInfo.OpenId,
			Valid:  u.WechatInfo.OpenId != "",
		},
		WechatUnionID: sql.NullString{
			String: u.WechatInfo.UnionId,
			Valid:  u.WechatInfo.UnionId != "",
		},
		Ctime: u.Ctime.UnixMilli(),
	}
}

func (r *CachedUserRepository) entityToDomain(u dao.User) domain.User {
	var birthday time.Time
	if u.Birthday.Valid {
		birthday = time.UnixMilli(u.Birthday.Int64)
	}
	return domain.User{
		Id:       u.Id,
		Email:    u.Email.String,
		Phone:    u.Phone.String,
		Password: u.Password,
		Nickname: u.Nickname.String,
		AboutMe:  u.AboutMe.String,
		WechatInfo: domain.WechatInfo{
			OpenId:  u.WechatOpenId.String,
			UnionId: u.WechatUnionID.String,
		},
		Birthday: birthday,
		Ctime:    time.UnixMilli(u.Ctime),
	}
}
