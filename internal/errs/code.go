package errs

const (
	// CommonInvalidInput 任何模块都可以使用的表达输入错误
	commonInvalidInput   = 400001
	commonInternalServer = 500001
)

const (
	// UserInvalidInput 用户模块输入错误，这是一个含糊的错误
	UserInvalidInput        = 401001
	UserInternalServerError = 501001
	// UserInvalidOrPassword 用户不存在或者密码错误
	UserInvalidOrPassword = 401002
)

const (
	ArticleInvalidInput        = 402001
	ArticleInternalServerError = 502001
)
