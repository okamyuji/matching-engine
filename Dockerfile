# ビルドステージ
FROM golang:1.25-alpine AS builder

# 必要なパッケージをインストール
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# 依存関係のキャッシュ
COPY go.mod go.sum ./
RUN go mod download

# ソースコードをコピー
COPY . .

# ビルド
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o /matching-engine \
    ./cmd/api

# 実行ステージ
FROM alpine:3.20

# セキュリティアップデート
RUN apk --no-cache add ca-certificates tzdata

# 非rootユーザー作成
RUN adduser -D -g '' appuser

WORKDIR /app

# バイナリをコピー
COPY --from=builder /matching-engine .

# 設定ファイルをコピー
COPY configs/ ./configs/

# 所有権を変更
RUN chown -R appuser:appuser /app

# 非rootユーザーで実行
USER appuser

# タイムゾーン設定
ENV TZ=Asia/Tokyo

# ポート公開
EXPOSE 8080

# ヘルスチェック
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health/live || exit 1

# エントリーポイント
CMD ["./matching-engine"]
