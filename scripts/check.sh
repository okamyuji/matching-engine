#!/usr/bin/env bash
# scripts/check.sh
# 品質チェック一括実行スクリプト（CI と同じゲート）
#
# 使用方法:
#   ./scripts/check.sh            # 全ゲート（mutation testing を含む）
#   SKIP_MUTATION=1 ./scripts/check.sh   # mutation testing を省略（開発中の高速確認）
#
# ゲート:
#   1. gofmt          2. go vet          3. staticcheck        4. golangci-lint
#   5. govulncheck    6. sqlc diff       7. コメント形式        8. go test（race, coverage）
#   9. カバレッジ >= 80%（sqlcgen / cmd / tools を除く）
#  10. CRAP < 15（tools/crap）
#  11. mutation testing（gremlins、効力 >= 80%、DB 非依存パッケージ）

set -euo pipefail

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

step() { echo -e "${YELLOW}[$1] $2${NC}"; }
ok() { echo -e "${GREEN}✓ $1${NC}\n"; }
fail() { echo -e "${RED}✗ $1${NC}"; exit 1; }

COVERAGE_MIN=${COVERAGE_MIN:-80}
CRAP_MAX=${CRAP_MAX:-15}

echo -e "${YELLOW}=== 品質チェック開始 ===${NC}\n"

step 1/11 "gofmt"
UNFMT=$(gofmt -l $(git ls-files '*.go' | grep -v '/sqlcgen/'))
[ -z "$UNFMT" ] || { echo "$UNFMT"; fail "フォーマットされていないファイルがあります"; }
ok "gofmt"

step 2/11 "go vet ./..."
go vet ./...
ok "vet"

step 3/11 "staticcheck ./..."
staticcheck ./...
ok "staticcheck"

step 4/11 "golangci-lint run ./..."
golangci-lint run ./...
ok "golangci-lint"

step 5/11 "govulncheck ./..."
govulncheck ./...
ok "govulncheck"

step 6/11 "sqlc diff（生成コードが最新か）"
sqlc diff
ok "sqlc"

step 7/11 "コメント形式（'// Name は ...' 形式を禁止、'// Name ...' に統一）"
BAD_COMMENTS=$(grep -rn --include='*.go' -E '^\s*// [A-Za-z_][A-Za-z0-9_]* は' . | grep -v '/sqlcgen/' || true)
[ -z "$BAD_COMMENTS" ] || { echo "$BAD_COMMENTS"; fail "「XXX はYYY」形式のコメントがあります。「XXX YYY」形式にしてください"; }
ok "コメント形式"

step 8/11 "go test -shuffle=on -count=1 -race -coverprofile=coverage.out ./..."
go test -shuffle=on -count=1 -race -coverprofile=coverage.out ./...
ok "test"

step 9/11 "カバレッジ >= ${COVERAGE_MIN}%（sqlcgen / cmd / tools を除く）"
grep -v -E '/sqlcgen/|/cmd/|/tools/' coverage.out > coverage.filtered.out
COVERAGE=$(go tool cover -func=coverage.filtered.out | grep total | awk '{print $3}' | sed 's/%//')
echo "Coverage: ${COVERAGE}%"
awk -v c="$COVERAGE" -v m="$COVERAGE_MIN" 'BEGIN { exit (c+0 >= m+0) ? 0 : 1 }' || fail "カバレッジ ${COVERAGE}% は ${COVERAGE_MIN}% 未満です"
ok "coverage"

step 10/11 "CRAP < ${CRAP_MAX}"
go run ./tools/crap -profile coverage.out -threshold "$CRAP_MAX"
ok "CRAP"

if [ "${SKIP_MUTATION:-0}" = "1" ]; then
  echo -e "${YELLOW}[11/11] mutation testing はスキップ（SKIP_MUTATION=1）${NC}\n"
else
  step 11/11 "mutation testing（gremlins）"
  gremlins unleash
  ok "mutation testing"
fi

echo -e "${GREEN}=== 全ゲート通過 ===${NC}"
