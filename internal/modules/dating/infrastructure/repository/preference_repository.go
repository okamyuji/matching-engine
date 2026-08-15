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

	if err := q.DeletePreferencePrefectures(ctx, pref.UserID); err != nil {
		return err
	}
	for _, p := range pref.Prefectures {
		if _, err := q.InsertPreferencePrefecture(ctx, sqlcgen.InsertPreferencePrefectureParams{UserID: pref.UserID, Prefecture: string(p.Prefecture)}); err != nil {
			return err
		}
	}
	if err := q.DeletePreferenceEducations(ctx, pref.UserID); err != nil {
		return err
	}
	for _, e := range pref.Educations {
		if _, err := q.InsertPreferenceEducation(ctx, sqlcgen.InsertPreferenceEducationParams{UserID: pref.UserID, Education: string(e.Education)}); err != nil {
			return err
		}
	}
	if err := q.DeletePreferenceMarriageDesires(ctx, pref.UserID); err != nil {
		return err
	}
	for _, m := range pref.MarriageDesires {
		if _, err := q.InsertPreferenceMarriageDesire(ctx, sqlcgen.InsertPreferenceMarriageDesireParams{UserID: pref.UserID, MarriageDesire: string(m.MarriageDesire)}); err != nil {
			return err
		}
	}
	if err := q.DeletePreferenceSmokingStatuses(ctx, pref.UserID); err != nil {
		return err
	}
	for _, s := range pref.SmokingStatuses {
		if _, err := q.InsertPreferenceSmokingStatus(ctx, sqlcgen.InsertPreferenceSmokingStatusParams{UserID: pref.UserID, SmokingStatus: string(s.SmokingStatus)}); err != nil {
			return err
		}
	}
	if err := q.DeletePreferenceDrinkingStatuses(ctx, pref.UserID); err != nil {
		return err
	}
	for _, d := range pref.DrinkingStatuses {
		if _, err := q.InsertPreferenceDrinkingStatus(ctx, sqlcgen.InsertPreferenceDrinkingStatusParams{UserID: pref.UserID, DrinkingStatus: string(d.DrinkingStatus)}); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}
