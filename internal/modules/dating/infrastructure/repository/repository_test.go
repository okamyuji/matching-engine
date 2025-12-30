package repository

import (
	"testing"
)

// 注記: データベース操作を伴う実際のリポジトリテストには結合テストが必要
// TestContainersを使用する。このテストはインターフェース準拠と構造体生成のみを検証する。

func TestLikeStruct(t *testing.T) {
	like := &Like{
		ID:         "like123",
		FromUserID: "user1",
		ToUserID:   "user2",
	}

	if like.ID == "" {
		t.Error("Like ID should not be empty")
	}
	if like.FromUserID == "" {
		t.Error("Like FromUserID should not be empty")
	}
	if like.ToUserID == "" {
		t.Error("Like ToUserID should not be empty")
	}
}

func TestUserWithProfile(t *testing.T) {
	uwp := &UserWithProfile{}

	// 構造体を作成でき、フィールドにアクセスできることを検証
	if uwp.User != nil {
		t.Error("User should be nil by default")
	}
	if uwp.Profile != nil {
		t.Error("Profile should be nil by default")
	}
}

func TestNewProfileRepository(t *testing.T) {
	repo := NewProfileRepository(nil)
	if repo == nil {
		t.Error("NewProfileRepository() returned nil")
	}
}

func TestNewLikeRepository(t *testing.T) {
	repo := NewLikeRepository(nil)
	if repo == nil {
		t.Error("NewLikeRepository() returned nil")
	}
}

func TestNewUserRepository(t *testing.T) {
	repo := NewUserRepository(nil)
	if repo == nil {
		t.Error("NewUserRepository() returned nil")
	}
}

func TestNewMatchRepository(t *testing.T) {
	repo := NewMatchRepository(nil)
	if repo == nil {
		t.Error("NewMatchRepository() returned nil")
	}
}
