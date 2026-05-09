package request

import "MathOverflow/services/forum/internal/model"

type SearchRequest struct {
	Query    string               `json:"query"`     // 搜索关键字
	Tags     []string             `json:"tags"`      // 标签过滤
	Page     int                  `json:"page"`      // 页码
	PageSize int                  `json:"page_size"` // 每页大小
	Sort     model.SearchSortType `json:"sort"`      // 排序字段
	From     int
}
