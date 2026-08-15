package application

import (
	"context"
	"errors"

	"github.com/okamyuji/matching-engine/internal/modules/dating/domain"
	"github.com/okamyuji/matching-engine/internal/modules/dating/infrastructure/repository"
)

// mockUserRepository ユーザーリポジトリのモック実装
type mockUserRepository struct {
	users       map[string]*domain.User
	profileRepo *mockProfileRepository
	err         error
}

func newMockUserRepository() *mockUserRepository {
	return &mockUserRepository{
		users: make(map[string]*domain.User),
	}
}

func (m *mockUserRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	if m.err != nil {
		return nil, m.err
	}
	user, ok := m.users[id]
	if !ok {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (m *mockUserRepository) FindCandidates(ctx context.Context, userID string, pref *domain.Preference) ([]*repository.UserWithProfile, error) {
	if m.err != nil {
		return nil, m.err
	}
	// Return some candidates if they exist in the users map
	results := make([]*repository.UserWithProfile, 0)
	for _, user := range m.users {
		if user.ID != userID {
			var profile *domain.Profile
			if m.profileRepo != nil {
				profile = m.profileRepo.getProfileForUser(user.ID)
			}
			results = append(results, &repository.UserWithProfile{
				User:    user,
				Profile: profile,
			})
		}
	}
	return results, nil
}

func (m *mockUserRepository) Create(ctx context.Context, user *domain.User) error {
	return m.err
}

func (m *mockUserRepository) Update(ctx context.Context, user *domain.User) error {
	return m.err
}

// mockProfileRepository プロフィールリポジトリのモック実装
type mockProfileRepository struct {
	profiles map[string]*domain.Profile
	err      error
}

func newMockProfileRepository() *mockProfileRepository {
	return &mockProfileRepository{
		profiles: make(map[string]*domain.Profile),
	}
}

func (m *mockProfileRepository) getProfileForUser(userID string) *domain.Profile {
	if profile, ok := m.profiles[userID]; ok {
		return profile
	}
	return nil
}

func (m *mockProfileRepository) FindByUserID(ctx context.Context, userID string) (*domain.Profile, error) {
	if m.err != nil {
		return nil, m.err
	}
	profile, ok := m.profiles[userID]
	if !ok {
		return nil, errors.New("profile not found")
	}
	return profile, nil
}

func (m *mockProfileRepository) Upsert(ctx context.Context, profile *domain.Profile) error {
	return m.err
}

// mockLikeRepository いいねリポジトリのモック実装
type mockLikeRepository struct {
	likes       []*repository.Like
	mutualCheck bool
	err         error
}

func newMockLikeRepository() *mockLikeRepository {
	return &mockLikeRepository{
		likes: make([]*repository.Like, 0),
	}
}

func (m *mockLikeRepository) Save(ctx context.Context, like *repository.Like) error {
	if m.err != nil {
		return m.err
	}
	m.likes = append(m.likes, like)
	return nil
}

func (m *mockLikeRepository) FindByTargetUserID(ctx context.Context, targetUserID string) ([]*repository.Like, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.likes, nil
}

func (m *mockLikeRepository) CheckMutual(ctx context.Context, userA, userB string) (bool, error) {
	if m.err != nil {
		return false, m.err
	}
	return m.mutualCheck, nil
}

func (m *mockLikeRepository) FindByFromUserID(ctx context.Context, fromUserID string) ([]*repository.Like, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.likes, nil
}

// mockMatchRepository マッチリポジトリのモック実装
type mockMatchRepository struct {
	matches []*domain.Match
	err     error
}

func newMockMatchRepository() *mockMatchRepository {
	return &mockMatchRepository{
		matches: make([]*domain.Match, 0),
	}
}

func (m *mockMatchRepository) Save(ctx context.Context, match *domain.Match) error {
	if m.err != nil {
		return m.err
	}
	m.matches = append(m.matches, match)
	return nil
}

func (m *mockMatchRepository) FindByUserID(ctx context.Context, userID string) ([]*domain.Match, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.matches, nil
}

func (m *mockMatchRepository) FindMutual(ctx context.Context, userID string) ([]*domain.Match, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.matches, nil
}
