package dao

import (
	"context"
	"github.com/ecodeclub/ekit/sqlx"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

var ErrWaitingSMSNotFound = gorm.ErrRecordNotFound

// AsyncSmsDao 异步短信
//
//go:generate mockgen.exe -source=./async_sms_dao.go -package=daomocks -destination=mocks/async_sms_dao.mock.go AsyncSmsDAO
type AsyncSmsDao interface {
	Insert(ctx context.Context, s AsyncSms) error
	GetWaitingSMS(ctx context.Context) (AsyncSms, error)
	MarkSuccess(ctx context.Context, id int64) error
	MarkFailed(ctx context.Context, id int64) error
}

const (
	asyncStatusWaiting = iota
	// 失败了，并且超过了重试次数
	asyncStatusFailed
	asyncStatusSuccess
)

type GORMAsyncSmsDao struct {
	db *gorm.DB
}

func NewGORMAsyncSmsDao(db *gorm.DB) AsyncSmsDao {
	return &GORMAsyncSmsDao{
		db: db,
	}
}

// Insert 写入 异步短信
func (dao *GORMAsyncSmsDao) Insert(ctx context.Context, s AsyncSms) error {
	now := time.Now().UnixMilli()
	s.Ctime = now
	return dao.db.WithContext(ctx).Create(&s).Error
}

// GetWaitingSMS 获取待发送短信
func (dao *GORMAsyncSmsDao) GetWaitingSMS(ctx context.Context) (AsyncSms, error) {
	// 如果在高并发情况下,SELECT for UPDATE 对数据库的压力很大
	// 但是我们不是高并发，因为你部署N台机器，才有 N 个goroutine 来查询
	// 并发不过百，随便写
	var s AsyncSms
	err := dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 为了避开一些偶发性的失败，我们只找 1 分钟前的异步短信发送
		now := time.Now().UnixMilli()
		endTime := now - time.Minute.Milliseconds()
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("utime < ? and status = ?",
				endTime, asyncStatusWaiting).First(&s).Error
		// SELECT xx FROM xxx WHERE xx FOR UPDATE，锁住了
		if err != nil {
			return err
		}

		// 只要更新了更新时间，根据我们前面的规则，就不可能被别的节点抢占了
		err = tx.Model(&AsyncSms{}).
			Where("id = ?", s.Id).
			Updates(map[string]any{
				"retry_cnt": gorm.Expr("retry_cnt + 1"),
				// 更新成了当前时间戳，确保我在发送过程中，没人会再次抢到它
				// 也相当于，重试间隔一分钟
				"utime": now,
			}).Error
		return err
	})
	return s, err
}

// MarkSuccess 标志成功
func (dao *GORMAsyncSmsDao) MarkSuccess(ctx context.Context, id int64) error {
	now := time.Now().UnixMilli()
	return dao.db.WithContext(ctx).Model(&AsyncSms{}).
		Where("id =?", id).
		Updates(map[string]any{
			"utime":  now,
			"status": asyncStatusSuccess,
		}).Error
}

// MarkFailed 标志失败
func (dao *GORMAsyncSmsDao) MarkFailed(ctx context.Context, id int64) error {
	now := time.Now().UnixMilli()
	return dao.db.WithContext(ctx).Model(&AsyncSms{}).
		// 只有到达了重试次数才会更新
		Where("id =? and `retry_cnt`>=`retry_max`", id).
		Updates(map[string]any{
			"utime":  now,
			"status": asyncStatusSuccess,
		}).Error

}

type AsyncSms struct {
	Id     int64
	Config sqlx.JsonColumn[SmsConfig]
	// 重试次数
	RetryCnt int `gorm:"column:retry_cnt"`
	// 最大的重试次数
	RetryMax int `gorm:"column:retry_max"`
	Status   uint8
	Ctime    int64
	Utime    int64 `gorm:"index"`
}

// SmsConfig 短信配置
type SmsConfig struct {
	// 短信模板id
	TplId string
	// 参数
	Args []string
	// 手机号
	Numbers []string
}
