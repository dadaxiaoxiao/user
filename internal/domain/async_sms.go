package domain

type AsyncSms struct {
	Id int64
	// 短信模板id
	TplId string
	// 参数
	Args []string
	// 手机号
	Numbers []string
	// 重试的配置
	RetryMax int
}
