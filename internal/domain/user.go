package domain

import "time"

// User  领域对象，DDD 中的entity
type User struct {
	Id       int64
	Email    string
	Nickname string
	Phone    string
	Password string
	AboutMe  string
	Ctime    time.Time
	Birthday time.Time
	// 如果将来接入 DingDingInfo，里面有同名字段 UnionID，所以不使用组合
	WechatInfo WechatInfo
}
