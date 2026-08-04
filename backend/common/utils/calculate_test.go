package utils

import (
	common "MathOverflow/common/model"
	"testing"
	"time"
)

func TestCalcPostScores_ZeroValues(t *testing.T) {
	post := &common.PostStat{
		Views:     0,
		Likes:     0,
		Stars:     0,
		Replies:   0,
		Status:    0,
		CreatedAt: time.Now(),
	}
	favors, score := CalcPostScores(post)
	if favors != 0 {
		t.Errorf("expected favors=0, got %d", favors)
	}
	if score != 0 {
		t.Errorf("expected score=0, got %f", score)
	}
}

func TestCalcPostScores_HighEngagement(t *testing.T) {
	post := &common.PostStat{
		Views:     1000,
		Likes:     50,
		Stars:     20,
		Replies:   30,
		Status:    1,
		CreatedAt: time.Now(),
	}
	favors, score := CalcPostScores(post)
	expectedFavors := int64(50*3 + 20*5) // 150+100=250
	if favors != expectedFavors {
		t.Errorf("expected favors=%d, got %d", expectedFavors, favors)
	}
	if score <= 0 {
		t.Errorf("expected positive score, got %f", score)
	}
}

func TestCalcPostScores_OlderPostScoresLower(t *testing.T) {
	base := &common.PostStat{
		Views:     100,
		Likes:     10,
		Stars:     5,
		Replies:   5,
		Status:    1,
		CreatedAt: time.Now(),
	}
	old := &common.PostStat{
		Views:     100,
		Likes:     10,
		Stars:     5,
		Replies:   5,
		Status:    1,
		CreatedAt: time.Now().Add(-48 * time.Hour),
	}
	_, baseScore := CalcPostScores(base)
	_, oldScore := CalcPostScores(old)
	if oldScore >= baseScore {
		t.Errorf("expected old score (%f) < base score (%f)", oldScore, baseScore)
	}
}

func TestCalcPostScores_FavorsFormula(t *testing.T) {
	post := &common.PostStat{
		Likes:     10,
		Stars:     4,
		CreatedAt: time.Now(),
	}
	favors, _ := CalcPostScores(post)
	expected := int64(10*3 + 4*5) // 30+20=50
	if favors != expected {
		t.Errorf("expected favors=%d, got %d", expected, favors)
	}
}
