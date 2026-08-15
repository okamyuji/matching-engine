package repository

import (
	"context"
	"fmt"

	"github.com/okamyuji/matching-engine/internal/modules/dating/domain"
	"github.com/okamyuji/matching-engine/internal/modules/dating/infrastructure/repository/sqlcgen"
)

// PreferenceRepository マッチング条件データアクセス用インターフェース
type PreferenceRepository interface {
	FindByUserID(ctx context.Context, userID string) (*domain.Preference, error)
	Upsert(ctx context.Context, pref *domain.Preference) error
}

// preferenceRepository PreferenceRepository の sqlc 実装
type preferenceRepository struct {
	db DB
	q  *sqlcgen.Queries
}

// NewPreferenceRepository 新しい PreferenceRepository を作成する
func NewPreferenceRepository(db DB) PreferenceRepository {
	return &preferenceRepository{db: db, q: sqlcgen.New(db)}
}

// FindByUserID ユーザーIDにより条件設定と詳細条件を取得する
func (r *preferenceRepository) FindByUserID(ctx context.Context, userID string) (*domain.Preference, error) {
	row, err := r.q.GetPreference(ctx, userID)
	if err != nil {
		return nil, err
	}
	pref := &domain.Preference{
		UserID:    row.UserID,
		AgeMin:    intFromInt32Ptr(row.AgeMin),
		AgeMax:    intFromInt32Ptr(row.AgeMax),
		HeightMin: intFromInt32Ptr(row.HeightMin),
		HeightMax: intFromInt32Ptr(row.HeightMax),
		IncomeMin: intFromInt32Ptr(row.IncomeMin),
		UpdatedAt: row.UpdatedAt,
	}

	prefs, err := r.q.ListPreferencePrefectures(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list prefectures: %w", err)
	}
	for _, p := range prefs {
		pref.Prefectures = append(pref.Prefectures, domain.PreferencePrefecture{ID: p.ID, UserID: p.UserID, Prefecture: domain.Prefecture(p.Prefecture)})
	}
	edus, err := r.q.ListPreferenceEducations(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list educations: %w", err)
	}
	for _, e := range edus {
		pref.Educations = append(pref.Educations, domain.PreferenceEducation{ID: e.ID, UserID: e.UserID, Education: domain.Education(e.Education)})
	}
	mds, err := r.q.ListPreferenceMarriageDesires(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list marriage desires: %w", err)
	}
	for _, m := range mds {
		pref.MarriageDesires = append(pref.MarriageDesires, domain.PreferenceMarriageDesire{ID: m.ID, UserID: m.UserID, MarriageDesire: domain.MarriageDesire(m.MarriageDesire)})
	}
	sss, err := r.q.ListPreferenceSmokingStatuses(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list smoking statuses: %w", err)
	}
	for _, s := range sss {
		pref.SmokingStatuses = append(pref.SmokingStatuses, domain.PreferenceSmokingStatus{ID: s.ID, UserID: s.UserID, SmokingStatus: domain.SmokingStatus(s.SmokingStatus)})
	}
	dss, err := r.q.ListPreferenceDrinkingStatuses(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list drinking statuses: %w", err)
	}
	for _, d := range dss {
		pref.DrinkingStatuses = append(pref.DrinkingStatuses, domain.PreferenceDrinkingStatus{ID: d.ID, UserID: d.UserID, DrinkingStatus: domain.DrinkingStatus(d.DrinkingStatus)})
	}
	return pref, nil
}

// Upsert 条件設定と詳細条件を1トランザクションで置き換える
func (r *preferenceRepository) Upsert(ctx context.Context, pref *domain.Preference) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	q := r.q.WithTx(tx)
	if err := q.UpsertPreference(ctx, sqlcgen.UpsertPreferenceParams{
		UserID:    pref.UserID,
		AgeMin:    int32PtrFromInt(pref.AgeMin),
		AgeMax:    int32PtrFromInt(pref.AgeMax),
		HeightMin: int32PtrFromInt(pref.HeightMin),
		HeightMax: int32PtrFromInt(pref.HeightMax),
		IncomeMin: int32PtrFromInt(pref.IncomeMin),
	}); err != nil {
		return err
	}
	for _, step := range preferenceChildSteps(q, pref) {
		if err := step(ctx); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

// preferenceChildSteps 詳細条件（都道府県・学歴・結婚意向・喫煙・飲酒）を
// 「全削除→再挿入」で置き換える手順を返す
func preferenceChildSteps(q *sqlcgen.Queries, pref *domain.Preference) []func(context.Context) error {
	uid := pref.UserID
	return []func(context.Context) error{
		func(ctx context.Context) error {
			return replaceChildren(ctx, func() error { return q.DeletePreferencePrefectures(ctx, uid) }, len(pref.Prefectures), func(i int) error {
				_, err := q.InsertPreferencePrefecture(ctx, sqlcgen.InsertPreferencePrefectureParams{UserID: uid, Prefecture: string(pref.Prefectures[i].Prefecture)})
				return err
			})
		},
		func(ctx context.Context) error {
			return replaceChildren(ctx, func() error { return q.DeletePreferenceEducations(ctx, uid) }, len(pref.Educations), func(i int) error {
				_, err := q.InsertPreferenceEducation(ctx, sqlcgen.InsertPreferenceEducationParams{UserID: uid, Education: string(pref.Educations[i].Education)})
				return err
			})
		},
		func(ctx context.Context) error {
			return replaceChildren(ctx, func() error { return q.DeletePreferenceMarriageDesires(ctx, uid) }, len(pref.MarriageDesires), func(i int) error {
				_, err := q.InsertPreferenceMarriageDesire(ctx, sqlcgen.InsertPreferenceMarriageDesireParams{UserID: uid, MarriageDesire: string(pref.MarriageDesires[i].MarriageDesire)})
				return err
			})
		},
		func(ctx context.Context) error {
			return replaceChildren(ctx, func() error { return q.DeletePreferenceSmokingStatuses(ctx, uid) }, len(pref.SmokingStatuses), func(i int) error {
				_, err := q.InsertPreferenceSmokingStatus(ctx, sqlcgen.InsertPreferenceSmokingStatusParams{UserID: uid, SmokingStatus: string(pref.SmokingStatuses[i].SmokingStatus)})
				return err
			})
		},
		func(ctx context.Context) error {
			return replaceChildren(ctx, func() error { return q.DeletePreferenceDrinkingStatuses(ctx, uid) }, len(pref.DrinkingStatuses), func(i int) error {
				_, err := q.InsertPreferenceDrinkingStatus(ctx, sqlcgen.InsertPreferenceDrinkingStatusParams{UserID: uid, DrinkingStatus: string(pref.DrinkingStatuses[i].DrinkingStatus)})
				return err
			})
		},
	}
}

// replaceChildren 子行を全削除してから n 件を挿入する
func replaceChildren(_ context.Context, del func() error, n int, insert func(i int) error) error {
	if err := del(); err != nil {
		return err
	}
	for i := 0; i < n; i++ {
		if err := insert(i); err != nil {
			return err
		}
	}
	return nil
}
