package testutil

import (
	"context"
	"testing"
)

func TestGetSharedTestDB_SeedAndClean(t *testing.T) {
	td := GetSharedTestDB(t)
	ctx := context.Background()

	td.SeedTestData(t)

	var users int
	if err := td.Pool.QueryRow(ctx, "SELECT count(*) FROM dating_users").Scan(&users); err != nil {
		t.Fatalf("count users: %v", err)
	}
	if users != 5 {
		t.Errorf("seeded users = %d, want 5", users)
	}

	td.CleanTables(t)
	if err := td.Pool.QueryRow(ctx, "SELECT count(*) FROM dating_users").Scan(&users); err != nil {
		t.Fatalf("count users after clean: %v", err)
	}
	if users != 0 {
		t.Errorf("users after clean = %d, want 0", users)
	}
}

func TestGetSharedTestDB_ReturnsSameInstance(t *testing.T) {
	a := GetSharedTestDB(t)
	b := GetSharedTestDB(t)
	if a != b {
		t.Fatal("GetSharedTestDB は同じインスタンスを返すべき")
	}
	if a.DBName == "" || a.DSN == "" {
		t.Fatalf("DBName/DSN が空: %+v", a)
	}
}

func TestReplaceDatabase(t *testing.T) {
	cases := map[string]string{
		"postgres://u:p@h:1/postgres?sslmode=disable": "postgres://u:p@h:1/x?sslmode=disable",
		"postgres://u:p@h:1/postgres":                 "postgres://u:p@h:1/x",
	}
	for in, want := range cases {
		if got := replaceDatabase(in, "x"); got != want {
			t.Errorf("replaceDatabase(%q) = %q, want %q", in, got, want)
		}
	}
}
