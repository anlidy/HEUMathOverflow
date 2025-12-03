package utils

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

func SliceFilter[T comparable](src []T, filter []T) []T {
	filterSet := make(map[T]struct{}, len(filter))
	for _, v := range filter {
		filterSet[v] = struct{}{}
	}
	res := make([]T, 0, len(src)) // 只分配容量
	for _, item := range src {
		if _, ok := filterSet[item]; !ok {
			res = append(res, item)
		}
	}
	return res
}

// IsPgDuplicateKey 判断是否为 PostgreSQL 唯一键冲突
func IsPgDuplicateKey(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}
