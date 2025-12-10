package model

type SearchSortType int

const (
	SortByFavors    SearchSortType = iota + 1 // 认可高(likes & stars)
	SortByViews                               // 浏览多
	SortByReplies                             // 评论多
	SortByCreatedAt                           // 新发布
	SortByDefault                             // 默认排序
)

func GetSortString(sort SearchSortType) string {
	switch sort {
	case 1:
		return "favors"
	case 2:
		return "views"
	case 3:
		return "replies"
	case 4:
		return "created_at"
	default:
		return "total_score"
	}
}
