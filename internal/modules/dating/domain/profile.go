package domain

import (
	"time"

	"github.com/uptrace/bun"
)

// Profile ユーザーのデートプロフィール
type Profile struct {
	bun.BaseModel `bun:"table:dating_profiles"`

	UserID           string         `bun:"user_id,pk"`
	Height           int            `bun:"height"`
	BodyType         BodyType       `bun:"body_type"`
	Education        Education      `bun:"education"`
	Occupation       string         `bun:"occupation"`
	IncomeLevel      int            `bun:"income_level"`
	MarriageDesire   MarriageDesire `bun:"marriage_desire"`
	ChildrenDesire   ChildrenDesire `bun:"children_desire"`
	Smoking          SmokingStatus  `bun:"smoking"`
	Drinking         DrinkingStatus `bun:"drinking"`
	SelfIntroduction string         `bun:"self_introduction"`
	UpdatedAt        time.Time      `bun:"updated_at,nullzero,default:current_timestamp"`

	// リレーション（Tags, Photosは別テーブルで管理）
	Tags   []ProfileTag   `bun:"rel:has-many,join:user_id=user_id"`
	Photos []ProfilePhoto `bun:"rel:has-many,join:user_id=user_id"`
}

// ProfileTag プロフィールタグ（dating_profile_tags テーブル）
type ProfileTag struct {
	bun.BaseModel `bun:"table:dating_profile_tags"`

	ID     int64  `bun:"id,pk,autoincrement"`
	UserID string `bun:"user_id,notnull"`
	Tag    string `bun:"tag,notnull"`
}

// ProfilePhoto プロフィール写真（dating_profile_photos テーブル）
type ProfilePhoto struct {
	bun.BaseModel `bun:"table:dating_profile_photos"`

	ID           int64     `bun:"id,pk,autoincrement"`
	UserID       string    `bun:"user_id,notnull"`
	URL          string    `bun:"url,notnull"`
	IsPrimary    bool      `bun:"is_primary,notnull,default:false"`
	DisplayOrder int       `bun:"display_order,notnull,default:0"`
	CreatedAt    time.Time `bun:"created_at,nullzero,default:current_timestamp"`
}
