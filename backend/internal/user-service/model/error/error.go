package syserror

// 定义错误类型
type Error int

const (
	NoError Error = iota
	EmailError
	PasswordError
	EmailExistsError
	NameExistsError
	NotFoundError
	InternalError
	DuplicateError
)
