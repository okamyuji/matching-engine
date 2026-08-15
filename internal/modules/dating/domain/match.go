package domain

import "time"

// Match 2人のユーザー間のマッチング結果
type Match struct {
	ID        string
	UserIDA   string
	UserIDB   string
	Score     float64
	Breakdown map[string]float64
	MatchedAt time.Time
}
