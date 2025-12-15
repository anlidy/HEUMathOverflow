package search

import (
	"MathOverflow/internal/forum-service/model/response"
	"encoding/json"
	"fmt"

	"github.com/elastic/go-elasticsearch/v8/esapi"
)

func ParseSearchResponse(res *esapi.Response) (*response.SearchResponse, error) {
	var raw struct {
		Hits struct {
			Total struct {
				Value int `json:"value"`
			} `json:"total"`
			Hits []struct {
				Source response.PostMeta `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(res.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("decode es response err: %w", err)
	}

	result := &response.SearchResponse{
		Total:     raw.Hits.Total.Value,
		PostMetas: make([]response.PostMeta, 0, len(raw.Hits.Hits)),
	}

	for _, hit := range raw.Hits.Hits {
		result.PostMetas = append(result.PostMetas, hit.Source)
	}
	return result, nil
}
