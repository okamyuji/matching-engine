# Universal Matching Engine

Modular Monolithアーキテクチャによる汎用マッチングエンジン

## クイックスタート

```bash
# 開発環境起動
docker compose up -d

# テスト実行
go test -shuffle=on -count=1 ./...

# 品質チェック
./scripts/check.sh

# ビルド
go build -o matching-engine ./cmd/api
```

## 技術スタック

| カテゴリ | 選択 |
| ------- | ---- |
| 言語 | Go 1.25+ |
| ORM | BUN |
| DB | MySQL 8.x |
| キャッシュ | Redis 7.x |
| コンテナ | Docker + docker compose |

## ドキュメント

Claude/AIでの設計・実装時は `CLAUDE.md` を参照してください。

| ファイル | 内容 |
| ------- | ---- |
| `CLAUDE.md` | エントリーポイント（AIアシスタント向け） |
| `claude/00_PROJECT_CONTEXT.md` | プロジェクトコンテキスト |
| `claude/01_ARCHITECTURE.md` | アーキテクチャ設計 |
| `claude/02_CORE_INTERFACES.md` | コアインターフェース |
| `claude/03_DOMAIN_DATING.md` | Datingドメイン詳細 |
| `claude/04_DOMAIN_MA.md` | M&Aドメイン詳細 |
| `claude/05_IMPLEMENTATION_GUIDE.md` | 実装ガイド |
| `claude/06_DATABASE_SCHEMA.md` | DB設計 |
| `claude/07_API_SPECIFICATION.md` | API仕様 |

## ディレクトリ構造

```text
matching-engine/
├── cmd/api/                 # エントリーポイント
├── internal/
│   ├── core/matching/       # コアエンジン
│   ├── modules/
│   │   ├── dating/          # Datingモジュール
│   │   └── ma/              # M&Aモジュール
│   └── shared/              # 共有インフラ
├── configs/                 # 設定ファイル
├── claude/                  # AI向けドキュメント
└── scripts/                 # ユーティリティ
```

## 開発

### 必須コマンド

```bash
go fmt ./...
go vet ./...
staticcheck ./...
golangci-lint run ./...
go test -shuffle=on -count=1 ./...
```

### カバレッジ

ユニットテストは**80%以上**のカバレッジを維持すること。

```bash
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
```

### Docker

```bash
# 起動
docker compose up -d

# ログ
docker compose logs -f app

# 停止
docker compose down
```

## 実装フェーズ

| Phase | 期間 | 内容 |
| ----- | ---- | ---- |
| 1 | 2週間 | コアエンジン |
| 2 | 2週間 | Datingモジュール |
| 3 | 2週間 | M&Aモジュール |
| 4 | 1週間 | インフラストラクチャ |

詳細は `claude/prompts/` を参照。

## ライセンス

MIT
