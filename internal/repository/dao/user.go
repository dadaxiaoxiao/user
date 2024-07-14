package dao

import (
	"context"
	"database/sql"
	"errors"
	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
	"time"
)

//go:generate mockgen.exe -source=./user.go -package=daomocks -destination=mocks/user.mock.go UserDao
type UserDao interface {
	Insert(ctx context.Context, u User) error
	FindByEmail(ctx context.Context, email string) (User, error)
	FindByPhone(ctx context.Context, phone string) (User, error)
	FindById(ctx context.Context, id int64) (User, error)
	UpdateNonZeroFields(ctx context.Context, u User) error
	FindByWechat(ctx context.Context, openID string) (User, error)
}

type GORMUserDAO struct {
	db *gorm.DB
}

var (
	ErrUserDuplicateEmail = errors.New("邮箱冲突")
	ErrUserNotFound       = gorm.ErrRecordNotFound
)

// NewGORMUserDAO 获取 结构实例
func NewGORMUserDAO(db *gorm.DB) UserDao {
	return &GORMUserDAO{
		db: db,
	}
}

// Insert 插入表
func (dao *GORMUserDAO) Insert(ctx context.Context, u User) error {
	// 当前毫秒
	now := time.Now().UnixMilli()
	u.Ctime = now
	u.Utime = now
	err := dao.db.WithContext(ctx).Create(&u).Error
	if mysqlErr, ok := err.(*mysql.MySQLError); ok {
		const uniqueConflictsErrNo uint16 = 1062
		if mysqlErr.Number == uniqueConflictsErrNo {
			// 邮箱冲突
			return ErrUserDuplicateEmail
		}
	}
	return err
}

// FindByEmail 根据email 查询用户信息
func (dao *GORMUserDAO) FindByEmail(ctx context.Context, email string) (User, error) {
	var u User
	err := dao.db.WithContext(ctx).Where("email = ?", email).First(&u).Error
	return u, err
}

// FindByPhone 根据Phone 查询用户信息
func (dao *GORMUserDAO) FindByPhone(ctx context.Context, phone string) (User, error) {
	var u User
	err := dao.db.WithContext(ctx).Where("phone = ?", phone).First(&u).Error
	return u, err
}

// FindByWechat 根据 微信凭据查找
func (dao *GORMUserDAO) FindByWechat(ctx context.Context, openID string) (User, error) {
	var u User
	err := dao.db.WithContext(ctx).Where("wechat_open_id = ?", openID).First(&u).Error
	return u, err
}

// FindById  根据id 查询用户信息
func (dao *GORMUserDAO) FindById(ctx context.Context, id int64) (User, error) {
	var u User
	err := dao.db.WithContext(ctx).Where("id = ?", id).First(&u).Error
	return u, err
}

// UpdateNonZeroFields 编辑信息
func (dao *GORMUserDAO) UpdateNonZeroFields(ctx context.Context, u User) error {
	now := time.Now().UnixMilli()
	u.Utime = now
	// 这种写法是很不清晰的，因为它依赖了 gorm 的两个默认语义
	// 会使用 ID 来作为 WHERE 条件
	// 会使用非零值来更新
	// 另外一种做法是显式指定只更新必要的字段，
	// 那么这意味着 DAO 和 service 中非敏感字段语义耦合了
	return dao.db.Updates(&u).Error
}

// User 数据库层次上的 用户表
type User struct {
	// 用户Id
	Id int64 `gorm:"primaryKey,autoIncrement"`
	// 邮箱
	Email sql.NullString `gorm:"colum:email;unique"`
	// 手机号
	// 唯一索引允许有多个空值 但是不能有多个 ""
	Phone sql.NullString `gorm:"colum:phone;unique"`
	// 密码
	Password string `gorm:"colum:password"`
	// 昵称
	Nickname sql.NullString `gorm:"colum:nickname"`
	// 生日
	Birthday sql.NullInt64 `gorm:"colum:birthday"`
	// 个人简介
	AboutMe sql.NullString `gorm:"colum:about_me;type:varchar(1024)"`
	// 微信Openid ,app 应用下唯一id
	WechatOpenId sql.NullString `gorm:"colum:wechat_openId;unique"`
	// 微信unionid
	WechatUnionID sql.NullString `gorm:"colum:wechat_unionID"`

	// 创建时间
	Ctime int64
	// 更新时间
	Utime int64
}
