package common

import "time"

type PostStat struct {
	Views      int64     `json:"views"`
	Likes      int64     `json:"likes"`
	Stars      int64     `json:"stars"`
	Replies    int64     `json:"replies"`
	Status     int       `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	Favors     int64     `json:"favors,omitempty"`
	TotalScore float64   `json:"total_score,omitempty"`
}
