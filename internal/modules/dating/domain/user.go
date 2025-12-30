package domain

import (
	"time"

	"github.com/uptrace/bun"
)

// User デートアプリのユーザーエンティティ
type User struct {
	bun.BaseModel `bun:"table:dating_users"`

	ID           string     `bun:"id,pk"`
	Nickname     string     `bun:"nickname,notnull"`
	Gender       Gender     `bun:"gender,notnull"`
	BirthDate    time.Time  `bun:"birth_date,notnull"`
	Prefecture   Prefecture `bun:"prefecture,notnull"`
	Verified     bool       `bun:"verified,notnull,default:false"`
	EloRating    int        `bun:"elo_rating,notnull,default:1000"`
	CreatedAt    time.Time  `bun:"created_at,nullzero,default:current_timestamp"`
	LastActiveAt time.Time  `bun:"last_active_at,nullzero,default:current_timestamp"`
}

// Age 生年月日からユーザーの年齢を計算する
func (u *User) Age() int {
	now := time.Now()
	age := now.Year() - u.BirthDate.Year()

	// 今年まだ誕生日が来ていない場合は調整
	if now.Month() < u.BirthDate.Month() ||
		(now.Month() == u.BirthDate.Month() && now.Day() < u.BirthDate.Day()) {
		age--
	}

	return age
}
