package model

type SearchSortType int

const (
	SortByDefault   SearchSortType = iota + 1 // 默认排序
	SortByFavors                              // 热度高(likes & stars)
	SortByCreatedAt                           // 新发布
	SortByViews                               // 浏览多
	SortByReplies                             // 评论多

)

func GetSortString(sort SearchSortType) string {
	switch sort {
	case 2:
		return "favors"
	case 3:
		return "created_at"
	case 4:
		return "views"
	case 5:
		return "replies"
	default:
		return "total_score"
	}
}
