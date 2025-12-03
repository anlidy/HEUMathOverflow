package model

// 排序类型
type OrderBy int

const (
	Recommend = iota // 推荐排序
	Hottest          // 热门排序
	Latest           // 最新排序
)
