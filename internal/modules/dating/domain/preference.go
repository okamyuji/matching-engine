package domain

import (
	"time"

	"github.com/uptrace/bun"
)

// Preference ユーザーのマッチング条件設定
type Preference struct {
	bun.BaseModel `bun:"table:dating_preferences"`

	UserID    string    `bun:"user_id,pk"`
	AgeMin    int       `bun:"age_min"`
	AgeMax    int       `bun:"age_max"`
	HeightMin int       `bun:"height_min"`
	HeightMax int       `bun:"height_max"`
	IncomeMin int       `bun:"income_min"`
	UpdatedAt time.Time `bun:"updated_at,nullzero,default:current_timestamp"`

	// リレーション（詳細条件は別テーブルで管理）
	Prefectures      []PreferencePrefecture     `bun:"rel:has-many,join:user_id=user_id"`
	Educations       []PreferenceEducation      `bun:"rel:has-many,join:user_id=user_id"`
	MarriageDesires  []PreferenceMarriageDesire `bun:"rel:has-many,join:user_id=user_id"`
	SmokingStatuses  []PreferenceSmokingStatus  `bun:"rel:has-many,join:user_id=user_id"`
	DrinkingStatuses []PreferenceDrinkingStatus `bun:"rel:has-many,join:user_id=user_id"`
}

// PreferencePrefecture 希望都道府県（dating_preference_prefectures テーブル）
type PreferencePrefecture struct {
	bun.BaseModel `bun:"table:dating_preference_prefectures"`

	ID         int64      `bun:"id,pk,autoincrement"`
	UserID     string     `bun:"user_id,notnull"`
	Prefecture Prefecture `bun:"prefecture,notnull"`
}

// PreferenceEducation 希望学歴（dating_preference_educations テーブル）
type PreferenceEducation struct {
	bun.BaseModel `bun:"table:dating_preference_educations"`

	ID        int64     `bun:"id,pk,autoincrement"`
	UserID    string    `bun:"user_id,notnull"`
	Education Education `bun:"education,notnull"`
}

// PreferenceMarriageDesire 希望結婚意向（dating_preference_marriage_desires テーブル）
type PreferenceMarriageDesire struct {
	bun.BaseModel `bun:"table:dating_preference_marriage_desires"`

	ID             int64          `bun:"id,pk,autoincrement"`
	UserID         string         `bun:"user_id,notnull"`
	MarriageDesire MarriageDesire `bun:"marriage_desire,notnull"`
}

// PreferenceSmokingStatus 希望喫煙状況（dating_preference_smoking_statuses テーブル）
type PreferenceSmokingStatus struct {
	bun.BaseModel `bun:"table:dating_preference_smoking_statuses"`

	ID            int64         `bun:"id,pk,autoincrement"`
	UserID        string        `bun:"user_id,notnull"`
	SmokingStatus SmokingStatus `bun:"smoking_status,notnull"`
}

// PreferenceDrinkingStatus 希望飲酒状況（dating_preference_drinking_statuses テーブル）
type PreferenceDrinkingStatus struct {
	bun.BaseModel `bun:"table:dating_preference_drinking_statuses"`

	ID             int64          `bun:"id,pk,autoincrement"`
	UserID         string         `bun:"user_id,notnull"`
	DrinkingStatus DrinkingStatus `bun:"drinking_status,notnull"`
}

// GetPrefectureStrings リレーションから都道府県文字列のスライスを取得する
func (p *Preference) GetPrefectureStrings() []string {
	result := make([]string, len(p.Prefectures))
	for i, pref := range p.Prefectures {
		result[i] = string(pref.Prefecture)
	}
	return result
}

// GetEducationStrings リレーションから学歴文字列のスライスを取得する
func (p *Preference) GetEducationStrings() []string {
	result := make([]string, len(p.Educations))
	for i, edu := range p.Educations {
		result[i] = string(edu.Education)
	}
	return result
}

// GetMarriageDesireStrings リレーションから結婚意向文字列のスライスを取得する
func (p *Preference) GetMarriageDesireStrings() []string {
	result := make([]string, len(p.MarriageDesires))
	for i, md := range p.MarriageDesires {
		result[i] = string(md.MarriageDesire)
	}
	return result
}

// GetSmokingStatusStrings リレーションから喫煙状況文字列のスライスを取得する
func (p *Preference) GetSmokingStatusStrings() []string {
	result := make([]string, len(p.SmokingStatuses))
	for i, ss := range p.SmokingStatuses {
		result[i] = string(ss.SmokingStatus)
	}
	return result
}

// GetDrinkingStatusStrings リレーションから飲酒状況文字列のスライスを取得する
func (p *Preference) GetDrinkingStatusStrings() []string {
	result := make([]string, len(p.DrinkingStatuses))
	for i, ds := range p.DrinkingStatuses {
		result[i] = string(ds.DrinkingStatus)
	}
	return result
}
