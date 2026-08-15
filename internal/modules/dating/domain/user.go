package domain

import "time"

// User デートアプリのユーザーエンティティ
type User struct {
	ID           string
	Nickname     string
	Gender       Gender
	BirthDate    time.Time
	Prefecture   Prefecture
	Verified     bool
	EloRating    int
	CreatedAt    time.Time
	LastActiveAt time.Time
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
