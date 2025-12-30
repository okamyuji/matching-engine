package domain

import (
	"time"

	"github.com/uptrace/bun"
)

// Match 2人のユーザー間のマッチング結果
type Match struct {
	bun.BaseModel `bun:"table:dating_matches"`

	ID        string             `bun:"id,pk"`
	UserIDA   string             `bun:"user_id_a,notnull"`
	UserIDB   string             `bun:"user_id_b,notnull"`
	Score     float64            `bun:"score,notnull"`
	Breakdown map[string]float64 `bun:"breakdown,type:json"`
	MatchedAt time.Time          `bun:"matched_at,nullzero,default:current_timestamp"`
}
