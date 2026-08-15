package domain

import "time"

// Preference ユーザーのマッチング条件設定
type Preference struct {
	UserID    string
	AgeMin    int
	AgeMax    int
	HeightMin int
	HeightMax int
	IncomeMin int
	UpdatedAt time.Time

	// リレーション（詳細条件は別テーブルで管理）
	Prefectures      []PreferencePrefecture
	Educations       []PreferenceEducation
	MarriageDesires  []PreferenceMarriageDesire
	SmokingStatuses  []PreferenceSmokingStatus
	DrinkingStatuses []PreferenceDrinkingStatus
}

// PreferencePrefecture 希望都道府県（dating_preference_prefectures テーブル）
type PreferencePrefecture struct {
	ID         int64
	UserID     string
	Prefecture Prefecture
}

// PreferenceEducation 希望学歴（dating_preference_educations テーブル）
type PreferenceEducation struct {
	ID        int64
	UserID    string
	Education Education
}

// PreferenceMarriageDesire 希望結婚意向（dating_preference_marriage_desires テーブル）
type PreferenceMarriageDesire struct {
	ID             int64
	UserID         string
	MarriageDesire MarriageDesire
}

// PreferenceSmokingStatus 希望喫煙状況（dating_preference_smoking_statuses テーブル）
type PreferenceSmokingStatus struct {
	ID            int64
	UserID        string
	SmokingStatus SmokingStatus
}

// PreferenceDrinkingStatus 希望飲酒状況（dating_preference_drinking_statuses テーブル）
type PreferenceDrinkingStatus struct {
	ID             int64
	UserID         string
	DrinkingStatus DrinkingStatus
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
