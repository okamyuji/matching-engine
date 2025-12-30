#!/usr/bin/env bash
# scripts/check.sh
# 品質チェック一括実行スクリプト
#
# 使用方法:
#   ./scripts/check.sh
#
# 必須コマンド:
#   go fmt/go vet/staticcheck/golangci-lint/go test
#   全てパスすること

set -e

# 色定義
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}=== 品質チェック開始 ===${NC}"
echo ""

# 1. go fmt
echo -e "${YELLOW}[1/6] go fmt ./...${NC}"
UNFMT=$(gofmt -l .)
if [ -n "$UNFMT" ]; then
    echo -e "${RED}フォーマットされていないファイル:${NC}"
    echo "$UNFMT"
    echo -e "${RED}go fmt ./... を実行してください${NC}"
    exit 1
fi
echo -e "${GREEN}✓ フォーマットOK${NC}"
echo ""

# 2. go vet
echo -e "${YELLOW}[2/6] go vet ./...${NC}"
go vet ./...
echo -e "${GREEN}✓ vet OK${NC}"
echo ""

# 3. staticcheck
echo -e "${YELLOW}[3/6] staticcheck ./...${NC}"
if command -v staticcheck &> /dev/null; then
    staticcheck ./...
    echo -e "${GREEN}✓ staticcheck OK${NC}"
else
    echo -e "${YELLOW}⚠ staticcheck がインストールされていません${NC}"
    echo "  go install honnef.co/go/tools/cmd/staticcheck@latest"
fi
echo ""

# 4. golangci-lint
echo -e "${YELLOW}[4/6] golangci-lint run ./...${NC}"
if command -v golangci-lint &> /dev/null; then
    golangci-lint run ./...
    echo -e "${GREEN}✓ golangci-lint OK${NC}"
else
    echo -e "${YELLOW}⚠ golangci-lint がインストールされていません${NC}"
    echo "  go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"
fi
echo ""

# 5. go test (shuffle + no cache + integration tests)
# cmdとtestutilパッケージを除外してテストを実行
echo -e "${YELLOW}[5/6] go test -tags=integration -shuffle=on -count=1 -coverprofile=coverage.out (excluding cmd/ and testutil/)${NC}"
go test -tags=integration -shuffle=on -count=1 -coverprofile=coverage.out $(go list ./... | grep -v "/cmd/" | grep -v "/testutil")
echo -e "${GREEN}✓ テスト OK${NC}"
echo ""

# 6. カバレッジチェック (>= 80%)
echo -e "${YELLOW}[6/6] カバレッジチェック (>= 80%)${NC}"

# coverage.outから直接totalを取得（すでにcmd/とtestutil/は除外済み）
COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')

if [ -z "$COVERAGE" ]; then
    echo -e "${YELLOW}⚠ カバレッジ情報がありません（テストがない可能性）${NC}"
else
    echo "Total coverage: ${COVERAGE}%"
    echo "(Note: Excludes cmd/ and testutil/ packages)"

    # 80%未満チェック
    if (( $(echo "$COVERAGE < 80" | bc -l) )); then
        echo -e "${RED}✗ カバレッジ ${COVERAGE}% は 80% 未満です${NC}"
        echo ""
        echo "主要パッケージのカバレッジ:"
        go test -tags=integration -coverprofile=coverage.out $(go list ./... | grep -v "/cmd/" | grep -v "/testutil") 2>&1 | grep "coverage:"
        exit 1
    fi
    echo -e "${GREEN}✓ カバレッジ OK (${COVERAGE}%)${NC}"
fi
echo ""

# 完了
echo -e "${GREEN}=== 全てのチェックが完了しました ===${NC}"
