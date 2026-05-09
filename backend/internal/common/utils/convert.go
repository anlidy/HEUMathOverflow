package utils

import (
	fmodel "MathOverflow/internal/forum-service/model"
	umodel "MathOverflow/internal/user-service/model"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// string -> objectID
func StringToObjectID(idStr string) (primitive.ObjectID, error) {
	return primitive.ObjectIDFromHex(idStr)
}

// objectID -> string
func ObjectIDToString(id primitive.ObjectID) string {
	return id.Hex()
}

// 获取role的字符串名称
func GetRoleName(role int) string {
	switch umodel.Role(role) {
	case umodel.Student:
		return "student"
	case umodel.Assistant:
		return "assitant"
	case umodel.Teacher:
		return "teacher"
	case umodel.Admin:
		return "admin"
	default:
		return ""
	}
}

// 获取PostStatus的字符串名称
func GetPostStatusName(role int) string {
	switch fmodel.AnswerStatus(role) {
	case fmodel.Unanswered:
		return "Unanswered"
	case fmodel.Answered:
		return "Answered"
	default:
		return ""
	}
}

// 获取ReplyStatus的字符串名称
func GetReplyStatusName(role int) string {
	switch fmodel.ReplyStatus(role) {
	case fmodel.NotSelected:
		return "NotSelected"
	case fmodel.AuthorSelected:
		return "AuthorSelected"
	case fmodel.TeacherCertified:
		return "TeacherCertified"
	default:
		return ""
	}
}
