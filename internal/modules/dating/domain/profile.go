package domain

import "time"

// Profile ユーザーのデートプロフィール
type Profile struct {
	UserID           string
	Height           int
	BodyType         BodyType
	Education        Education
	Occupation       string
	IncomeLevel      int
	MarriageDesire   MarriageDesire
	ChildrenDesire   ChildrenDesire
	Smoking          SmokingStatus
	Drinking         DrinkingStatus
	SelfIntroduction string
	UpdatedAt        time.Time

	// リレーション（Tags, Photosは別テーブルで管理）
	Tags   []ProfileTag
	Photos []ProfilePhoto
}

// ProfileTag プロフィールタグ（dating_profile_tags テーブル）
type ProfileTag struct {
	ID     int64
	UserID string
	Tag    string
}

// ProfilePhoto プロフィール写真（dating_profile_photos テーブル）
type ProfilePhoto struct {
	ID           int64
	UserID       string
	URL          string
	IsPrimary    bool
	DisplayOrder int
	CreatedAt    time.Time
}
