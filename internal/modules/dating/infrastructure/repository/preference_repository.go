package repository

import (
	"context"

	"github.com/uptrace/bun"
	"github.com/yourorg/matching-engine/internal/modules/dating/domain"
)

// PreferenceRepository 希望条件データアクセス用インターフェース
type PreferenceRepository interface {
	FindByUserID(ctx context.Context, userID string) (*domain.Preference, error)
	Upsert(ctx context.Context, pref *domain.Preference) error
}

// preferenceRepository PreferenceRepositoryのBUN実装
type preferenceRepository struct {
	db *bun.DB
}

// NewPreferenceRepository 新しいPreferenceRepositoryを作成する
func NewPreferenceRepository(db *bun.DB) PreferenceRepository {
	return &preferenceRepository{db: db}
}

// FindByUserID ユーザーIDで希望条件を取得する（関連データも含む）
func (r *preferenceRepository) FindByUserID(ctx context.Context, userID string) (*domain.Preference, error) {
	pref := &domain.Preference{}

	err := r.db.NewSelect().
		Model(pref).
		Relation("Prefectures").
		Relation("Educations").
		Relation("MarriageDesires").
		Relation("SmokingStatuses").
		Relation("DrinkingStatuses").
		Where("user_id = ?", userID).
		Scan(ctx)

	if err != nil {
		return nil, err
	}

	return pref, nil
}

// Upsert 希望条件を挿入または更新する
func (r *preferenceRepository) Upsert(ctx context.Context, pref *domain.Preference) error {
	// メインテーブルのUpsert
	_, err := r.db.NewInsert().
		Model(pref).
		On("DUPLICATE KEY UPDATE").
		Set("age_min = VALUES(age_min)").
		Set("age_max = VALUES(age_max)").
		Set("height_min = VALUES(height_min)").
		Set("height_max = VALUES(height_max)").
		Set("income_min = VALUES(income_min)").
		Set("updated_at = CURRENT_TIMESTAMP").
		Exec(ctx)

	if err != nil {
		return err
	}

	// 関連テーブルの更新（既存を削除してから挿入）
	if err := r.updatePrefectures(ctx, pref.UserID, pref.Prefectures); err != nil {
		return err
	}

	if err := r.updateEducations(ctx, pref.UserID, pref.Educations); err != nil {
		return err
	}

	if err := r.updateMarriageDesires(ctx, pref.UserID, pref.MarriageDesires); err != nil {
		return err
	}

	if err := r.updateSmokingStatuses(ctx, pref.UserID, pref.SmokingStatuses); err != nil {
		return err
	}

	if err := r.updateDrinkingStatuses(ctx, pref.UserID, pref.DrinkingStatuses); err != nil {
		return err
	}

	return nil
}

// updatePrefectures 都道府県設定を更新する
func (r *preferenceRepository) updatePrefectures(ctx context.Context, userID string, prefectures []domain.PreferencePrefecture) error {
	// 既存データ削除
	_, err := r.db.NewDelete().
		Model((*domain.PreferencePrefecture)(nil)).
		Where("user_id = ?", userID).
		Exec(ctx)

	if err != nil {
		return err
	}

	// 新規データ挿入
	if len(prefectures) > 0 {
		_, err = r.db.NewInsert().
			Model(&prefectures).
			Exec(ctx)
	}

	return err
}

// updateEducations 学歴設定を更新する
func (r *preferenceRepository) updateEducations(ctx context.Context, userID string, educations []domain.PreferenceEducation) error {
	// 既存データ削除
	_, err := r.db.NewDelete().
		Model((*domain.PreferenceEducation)(nil)).
		Where("user_id = ?", userID).
		Exec(ctx)

	if err != nil {
		return err
	}

	// 新規データ挿入
	if len(educations) > 0 {
		_, err = r.db.NewInsert().
			Model(&educations).
			Exec(ctx)
	}

	return err
}

// updateMarriageDesires 結婚意向設定を更新する
func (r *preferenceRepository) updateMarriageDesires(ctx context.Context, userID string, desires []domain.PreferenceMarriageDesire) error {
	// 既存データ削除
	_, err := r.db.NewDelete().
		Model((*domain.PreferenceMarriageDesire)(nil)).
		Where("user_id = ?", userID).
		Exec(ctx)

	if err != nil {
		return err
	}

	// 新規データ挿入
	if len(desires) > 0 {
		_, err = r.db.NewInsert().
			Model(&desires).
			Exec(ctx)
	}

	return err
}

// updateSmokingStatuses 喫煙状況設定を更新する
func (r *preferenceRepository) updateSmokingStatuses(ctx context.Context, userID string, statuses []domain.PreferenceSmokingStatus) error {
	// 既存データ削除
	_, err := r.db.NewDelete().
		Model((*domain.PreferenceSmokingStatus)(nil)).
		Where("user_id = ?", userID).
		Exec(ctx)

	if err != nil {
		return err
	}

	// 新規データ挿入
	if len(statuses) > 0 {
		_, err = r.db.NewInsert().
			Model(&statuses).
			Exec(ctx)
	}

	return err
}

// updateDrinkingStatuses 飲酒状況設定を更新する
func (r *preferenceRepository) updateDrinkingStatuses(ctx context.Context, userID string, statuses []domain.PreferenceDrinkingStatus) error {
	// 既存データ削除
	_, err := r.db.NewDelete().
		Model((*domain.PreferenceDrinkingStatus)(nil)).
		Where("user_id = ?", userID).
		Exec(ctx)

	if err != nil {
		return err
	}

	// 新規データ挿入
	if len(statuses) > 0 {
		_, err = r.db.NewInsert().
			Model(&statuses).
			Exec(ctx)
	}

	return err
}
