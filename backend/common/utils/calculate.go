package utils

import (
	common "MathOverflow/common/model"
	"math"
	"time"
)

func CalcPostScores(post *common.PostStat) (int64, float64) {
	hours := time.Since(post.CreatedAt).Seconds()/3600 + 2
	timeFactor := math.Pow(hours, 1.5)
	Favors := post.Likes*3 + post.Stars*5
	TotalScore := float64(post.Views/10+post.Likes*3+post.Replies*4+post.Stars*5+int64(post.Status*10)) / float64(timeFactor)
	return Favors, TotalScore
}
