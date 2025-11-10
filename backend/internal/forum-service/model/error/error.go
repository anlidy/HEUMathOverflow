package syserror

// 定义错误类型
type Error int

const (
	NoError Error = iota
	NotFoundError
	InternalError
	DuplicateError
	PermissionDeniedError
	NetworkError
	ResourceExpiredError
)
