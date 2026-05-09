package search

import (
	"MathOverflow/services/forum/internal/model"
	"MathOverflow/services/forum/internal/model/request"
)

// 构建 ES 查询 DSL
func BuildQuery(req request.SearchRequest) map[string]any {
	order := "desc" // 都采用降序
	sortBy := model.GetSortString(req.Sort)

	must := []map[string]any{}
	filter := []map[string]any{}

	// 关键词搜索
	if req.Query != "" {
		must = append(must, map[string]any{
			"multi_match": map[string]any{
				"query":  req.Query,                                       // 基于分词的模糊匹配
				"fields": []string{"title^3", "author_name^2", "content"}, // 匹配区域及权重
			},
		})
	}

	// Tag 过滤
	if len(req.Tags) > 0 {
		filter = append(filter, map[string]any{
			"terms": map[string]any{
				"tags.keyword": req.Tags, // 关键词匹配,不分词
			},
		})
	}

	// 排序
	sort := []map[string]any{
		{
			sortBy: map[string]any{
				"order": order,
			},
		},
	}

	// 构建body
	body := map[string]any{
		"query": map[string]any{
			"bool": map[string]any{
				"must":   must,
				"filter": filter,
			},
		},
		"sort": sort,
		"from": req.From,
		"size": req.PageSize,
	}

	return body
}
