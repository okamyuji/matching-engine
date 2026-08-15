// Package e2e HTTP API の主要導線を、実際の PostgreSQL（testcontainers）と
// 本番と同じ配線（internal/app）で end-to-end に検証する。
package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/okamyuji/matching-engine/internal/app"
	maDomain "github.com/okamyuji/matching-engine/internal/modules/ma/domain"
	maRepo "github.com/okamyuji/matching-engine/internal/modules/ma/infrastructure/repository"
	"github.com/okamyuji/matching-engine/internal/shared/auth"
	"github.com/okamyuji/matching-engine/internal/testutil"
)

type env struct {
	t      *testing.T
	server *httptest.Server
	td     *testutil.TestDatabase
}

func newEnv(t *testing.T) *env {
	t.Helper()
	td := testutil.GetSharedTestDB(t)
	td.SeedTestData(t)

	root := repoRoot(t)
	handler, err := app.NewRouter(td.Pool, app.Options{
		DatingConfigPath: filepath.Join(root, "configs", "dating", "matching.json"),
		MAConfigPath:     filepath.Join(root, "configs", "ma", "matching.json"),
	})
	if err != nil {
		t.Fatalf("NewRouter: %v", err)
	}
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return &env{t: t, server: srv, td: td}
}

func repoRoot(t *testing.T) string {
	t.Helper()
	// test/e2e から2つ上がリポジトリルート
	abs, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("repo root: %v", err)
	}
	return abs
}

func userToken(t *testing.T, userID string) string {
	t.Helper()
	tok, err := auth.Sign(auth.SecretFromEnv(), map[string]any{"user_id": userID, "exp": time.Now().Add(time.Hour).Unix()})
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	return tok
}

func companyToken(t *testing.T, companyID string) string {
	t.Helper()
	tok, err := auth.Sign(auth.SecretFromEnv(), map[string]any{"company_id": companyID, "exp": time.Now().Add(time.Hour).Unix()})
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	return tok
}

// do リクエストを送り、ステータスと本文を返す
func (e *env) do(method, path, token string, body any) (int, []byte) {
	e.t.Helper()
	var reader io.Reader
	if body != nil {
		if raw, ok := body.([]byte); ok {
			reader = bytes.NewReader(raw)
		} else {
			b, err := json.Marshal(body)
			if err != nil {
				e.t.Fatalf("marshal body: %v", err)
			}
			reader = bytes.NewReader(b)
		}
	}
	req, err := http.NewRequestWithContext(context.Background(), method, e.server.URL+path, reader)
	if err != nil {
		e.t.Fatalf("new request: %v", err)
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := e.server.Client().Do(req)
	if err != nil {
		e.t.Fatalf("%s %s: %v", method, path, err)
	}
	defer func() { _ = resp.Body.Close() }()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		e.t.Fatalf("read body: %v", err)
	}
	return resp.StatusCode, data
}

func mustJSON(t *testing.T, data []byte, v any) {
	t.Helper()
	if err := json.Unmarshal(data, v); err != nil {
		t.Fatalf("unmarshal %q: %v", string(data), err)
	}
}

// ---- ヘルスチェック ----

func TestHealth(t *testing.T) {
	e := newEnv(t)
	for _, path := range []string{"/health/live", "/health/ready"} {
		status, body := e.do(http.MethodGet, path, "", nil)
		if status != http.StatusOK {
			t.Errorf("%s status = %d body=%s", path, status, body)
		}
	}
}

// ---- Dating 正常導線 ----

func TestDating_HappyPath(t *testing.T) {
	e := newEnv(t)
	tok1 := userToken(t, "user1")
	tok2 := userToken(t, "user2")

	// 候補取得（シードの user1 は設定あり）
	status, body := e.do(http.MethodGet, "/api/v1/dating/matches?limit=5", tok1, nil)
	if status != http.StatusOK {
		t.Fatalf("matches status = %d body=%s", status, body)
	}
	var matches []map[string]any
	mustJSON(t, body, &matches)
	if len(matches) == 0 {
		t.Fatalf("候補が0件: %s", body)
	}
	if len(matches) > 5 {
		t.Errorf("limit が効いていない: %d", len(matches))
	}

	// user1 → user2 いいね（まだマッチしない）
	status, body = e.do(http.MethodPost, "/api/v1/dating/likes", tok1, map[string]string{"target_user_id": "user2"})
	if status != http.StatusOK {
		t.Fatalf("like status = %d body=%s", status, body)
	}
	var like map[string]any
	mustJSON(t, body, &like)
	if like["matched"] != false {
		t.Errorf("片思いでマッチしてはいけない: %v", like)
	}

	// user2 が受け取ったいいねに user1 がいる
	status, body = e.do(http.MethodGet, "/api/v1/dating/likes/received", tok2, nil)
	if status != http.StatusOK {
		t.Fatalf("received status = %d body=%s", status, body)
	}
	var received []map[string]any
	mustJSON(t, body, &received)
	if len(received) != 1 || received[0]["FromUserID"] != "user1" && received[0]["from_user_id"] != "user1" {
		t.Errorf("受信いいねが想定外: %s", body)
	}

	// user2 → user1 いいねで相互マッチ成立
	status, body = e.do(http.MethodPost, "/api/v1/dating/likes", tok2, map[string]string{"target_user_id": "user1"})
	if status != http.StatusOK {
		t.Fatalf("like back status = %d body=%s", status, body)
	}
	mustJSON(t, body, &like)
	if like["matched"] != true || like["match_id"] == "" {
		t.Errorf("相互いいねでマッチすべき: %v", like)
	}

	// 相互マッチ一覧に現れる
	status, body = e.do(http.MethodGet, "/api/v1/dating/matches/mutual", tok1, nil)
	if status != http.StatusOK {
		t.Fatalf("mutual status = %d body=%s", status, body)
	}
	var mutual []map[string]any
	mustJSON(t, body, &mutual)
	if len(mutual) != 1 {
		t.Errorf("相互マッチは1件のはず: %s", body)
	}
}

// ---- Dating 異常導線 ----

func TestDating_ErrorPaths(t *testing.T) {
	e := newEnv(t)
	tok1 := userToken(t, "user1")

	cases := []struct {
		name   string
		method string
		path   string
		token  string
		body   any
		want   int
	}{
		{"認証なし", http.MethodGet, "/api/v1/dating/matches", "", nil, http.StatusUnauthorized},
		{"不正トークン", http.MethodGet, "/api/v1/dating/matches", "not.a.token", nil, http.StatusUnauthorized},
		{"不正JSON", http.MethodPost, "/api/v1/dating/likes", tok1, []byte("{"), http.StatusBadRequest},
		{"target_user_id 欠落", http.MethodPost, "/api/v1/dating/likes", tok1, map[string]string{}, http.StatusBadRequest},
		{"存在しないユーザーの候補取得", http.MethodGet, "/api/v1/dating/matches", userToken(t, "nobody"), nil, http.StatusInternalServerError},
		{"メソッド不一致", http.MethodDelete, "/api/v1/dating/likes", tok1, nil, http.StatusMethodNotAllowed},
		{"存在しないパス", http.MethodGet, "/api/v1/dating/unknown", tok1, nil, http.StatusNotFound},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			status, body := e.do(tc.method, tc.path, tc.token, tc.body)
			if status != tc.want {
				t.Errorf("status = %d, want %d body=%s", status, tc.want, body)
			}
		})
	}

	// 期限切れトークン
	expired, _ := auth.Sign(auth.SecretFromEnv(), map[string]any{"user_id": "user1", "exp": time.Now().Add(-time.Minute).Unix()})
	if status, _ := e.do(http.MethodGet, "/api/v1/dating/matches", expired, nil); status != http.StatusUnauthorized {
		t.Errorf("期限切れトークン status = %d", status)
	}

	// 同じ相手への二重いいねは一意制約で失敗する
	if status, _ := e.do(http.MethodPost, "/api/v1/dating/likes", tok1, map[string]string{"target_user_id": "user3"}); status != http.StatusOK {
		t.Fatalf("初回いいね status = %d", status)
	}
	if status, _ := e.do(http.MethodPost, "/api/v1/dating/likes", tok1, map[string]string{"target_user_id": "user3"}); status != http.StatusInternalServerError {
		t.Errorf("二重いいね status = %d, want 500", status)
	}
}

// ---- M&A 正常導線 ----

func seedMA(t *testing.T, e *env) {
	t.Helper()
	ctx := context.Background()
	companies := maRepo.NewCompanyRepository(e.td.Pool)
	financials := maRepo.NewFinancialsRepository(e.td.Pool)
	founded := time.Date(2010, 1, 1, 0, 0, 0, 0, time.UTC)

	newCompany := func(id, name string, industry maDomain.Industry, purpose maDomain.MatchingPurpose, employees int) {
		c := &maDomain.Company{ID: id, Name: name, Industry: industry, Location: "Tokyo", EmployeeCount: employees, Founded: founded, ListingStatus: maDomain.ListingPrivate, MatchingPurpose: purpose}
		if err := companies.Create(ctx, c); err != nil {
			t.Fatalf("create company %s: %v", id, err)
		}
		if err := maRepo.InsertTechnology(ctx, e.td.Pool, &maDomain.CompanyTechnology{CompanyID: id, Technology: "cloud"}); err != nil {
			t.Fatalf("technology %s: %v", id, err)
		}
		if err := maRepo.InsertMarket(ctx, e.td.Pool, &maDomain.CompanyMarket{CompanyID: id, Market: "japan"}); err != nil {
			t.Fatalf("market %s: %v", id, err)
		}
		for i, year := range []int{2022, 2023, 2024} {
			f := &maDomain.Financials{CompanyID: id, FiscalYear: year, Revenue: int64(1_000_000_000 * (i + 1)), EBITDA: int64(100_000_000 * (i + 1)), NetIncome: 50_000_000, TotalAssets: 2_000_000_000, TotalLiabilities: 800_000_000, Equity: 1_200_000_000, ROE: 0.1, ROA: 0.05, DebtEquityRatio: 0.7, CurrentRatio: 1.5}
			if err := financials.Save(ctx, f); err != nil {
				t.Fatalf("financials %s: %v", id, err)
			}
		}
	}
	newCompany("buyer1", "Buyer Inc", maDomain.IndustryTechnology, maDomain.PurposeBuyer, 500)
	newCompany("seller1", "Seller One", maDomain.IndustryTechnology, maDomain.PurposeSeller, 120)
	newCompany("seller2", "Seller Two", maDomain.IndustryFinance, maDomain.PurposeSeller, 80)

	if err := maRepo.UpsertCriteria(ctx, e.td.Pool, &maDomain.MAMatchingCriteria{
		CompanyID: "buyer1", Purpose: maDomain.PurposeBuyer, EmployeeMin: 10, EmployeeMax: 1000,
		TargetIndustries: []maDomain.CriteriaIndustry{{CompanyID: "buyer1", Industry: maDomain.IndustryTechnology}, {CompanyID: "buyer1", Industry: maDomain.IndustryFinance}},
	}); err != nil {
		t.Fatalf("criteria: %v", err)
	}
}

func TestMA_HappyPath(t *testing.T) {
	e := newEnv(t)
	seedMA(t, e)
	buyer := companyToken(t, "buyer1")
	seller := companyToken(t, "seller1")

	status, body := e.do(http.MethodGet, "/api/v1/ma/targets?limit=10", buyer, nil)
	if status != http.StatusOK {
		t.Fatalf("targets status = %d body=%s", status, body)
	}
	var targets []map[string]any
	mustJSON(t, body, &targets)
	if len(targets) != 2 {
		t.Fatalf("候補は seller1, seller2 の2件のはず: %s", body)
	}

	status, body = e.do(http.MethodPost, "/api/v1/ma/interests", buyer, map[string]string{"target_company_id": "seller1"})
	if status != http.StatusOK {
		t.Fatalf("interest status = %d body=%s", status, body)
	}
	var interest map[string]any
	mustJSON(t, body, &interest)
	if interest["matched"] != false {
		t.Errorf("片側の関心表明でマッチしてはいけない: %v", interest)
	}

	status, body = e.do(http.MethodGet, "/api/v1/ma/interests/received", seller, nil)
	if status != http.StatusOK {
		t.Fatalf("received status = %d body=%s", status, body)
	}
	var received []map[string]any
	mustJSON(t, body, &received)
	if len(received) != 1 {
		t.Errorf("受信した関心表明は1件のはず: %s", body)
	}

	status, body = e.do(http.MethodPost, "/api/v1/ma/interests", seller, map[string]string{"target_company_id": "buyer1"})
	if status != http.StatusOK {
		t.Fatalf("interest back status = %d body=%s", status, body)
	}
	mustJSON(t, body, &interest)
	if interest["matched"] != true {
		t.Errorf("相互の関心表明でマッチすべき: %v", interest)
	}

	status, body = e.do(http.MethodGet, "/api/v1/ma/matches", buyer, nil)
	if status != http.StatusOK {
		t.Fatalf("matches status = %d body=%s", status, body)
	}
	var matches []map[string]any
	mustJSON(t, body, &matches)
	if len(matches) != 1 {
		t.Errorf("相互マッチは1件のはず: %s", body)
	}
}

// ---- M&A 異常導線 ----

func TestMA_ErrorPaths(t *testing.T) {
	e := newEnv(t)
	seedMA(t, e)
	buyer := companyToken(t, "buyer1")

	cases := []struct {
		name   string
		method string
		path   string
		token  string
		body   any
		want   int
	}{
		{"認証なし", http.MethodGet, "/api/v1/ma/targets", "", nil, http.StatusUnauthorized},
		{"user_id クレームの誤ったトークン", http.MethodGet, "/api/v1/ma/targets", userToken(t, "user1"), nil, http.StatusUnauthorized},
		{"不正JSON", http.MethodPost, "/api/v1/ma/interests", buyer, []byte("not json"), http.StatusBadRequest},
		{"target_company_id 欠落", http.MethodPost, "/api/v1/ma/interests", buyer, map[string]string{}, http.StatusBadRequest},
		{"存在しない企業の候補取得", http.MethodGet, "/api/v1/ma/targets", companyToken(t, "ghost"), nil, http.StatusInternalServerError},
		{"バリュエーションは未実装", http.MethodGet, "/api/v1/ma/valuation/buyer1", buyer, nil, http.StatusNotImplemented},
		{"メソッド不一致", http.MethodPut, "/api/v1/ma/interests", buyer, nil, http.StatusMethodNotAllowed},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			status, body := e.do(tc.method, tc.path, tc.token, tc.body)
			if status != tc.want {
				t.Errorf("status = %d, want %d body=%s", status, tc.want, body)
			}
		})
	}

	// 同じ相手への二重の関心表明はサービス層で拒否される
	if status, _ := e.do(http.MethodPost, "/api/v1/ma/interests", buyer, map[string]string{"target_company_id": "seller2"}); status != http.StatusOK {
		t.Fatalf("初回関心表明 status = %d", status)
	}
	if status, _ := e.do(http.MethodPost, "/api/v1/ma/interests", buyer, map[string]string{"target_company_id": "seller2"}); status != http.StatusInternalServerError {
		t.Errorf("二重の関心表明 status = %d, want 500", status)
	}
}
