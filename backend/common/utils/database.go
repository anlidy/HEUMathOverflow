package utils

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// IsPgDuplicateKey 判断是否为 PostgreSQL 唯一键冲突
func IsPgDuplicateKey(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}

// // 判断是否为外键约束错误
func IsPgViolateForeignKey(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23503"
	}
	return false
}

// string -> objectID
func StringToObjectID(idStr string) (primitive.ObjectID, error) {
	return primitive.ObjectIDFromHex(idStr)
}

// objectID -> string
func ObjectIDToString(id primitive.ObjectID) string {
	return id.Hex()
}
