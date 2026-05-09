package common

type ReplyStatus int

const (
	NotSelected      ReplyStatus = iota + 1
	AuthorSelected               // 作者选定的答案
	TeacherCertified             // 教师精选的答案
)
