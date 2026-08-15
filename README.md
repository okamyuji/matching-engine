# Universal Matching Engine

Modular Monolithアーキテクチャによる汎用マッチングエンジンです。設定ファイル駆動の重み付きスコアリング（コアエンジン）の上に、Dating と M&A の2つのドメインモジュールを載せています。

## 初版からの変更点

初版（BUN + MySQL 8 + Redis）から、次の点を変更しました。骨子（Modular Monolith、コアエンジン、Dating / M&A モジュール、API）は変えていません。

| 領域 | 初版 | 現在 |
| ---- | ---- | ---- |
| データベース | MySQL 8.x | PostgreSQL 18（列挙値は `text` + `CHECK`、主キー採番は identity、時刻は `timestamptz`、JSON は `jsonb`） |
| データアクセス | BUN（クエリビルダ、構造体タグ） | sqlc（`queries/*.sql` から型付き Go コードを生成、`pgx/v5`） |
| キャッシュ | Redis（設定のみで未使用） | 撤去 |
| ドメイン構造体 | BUN タグ付き | タグなしの純粋な構造体 |
| モジュールパス | `github.com/yourorg/matching-engine` | `github.com/okamyuji/matching-engine` |
| テスト DB | テストごとに MySQL コンテナ | PostgreSQL 18 コンテナを名前付きで再利用し、テンプレート DB からプロセス別 DB を払い出す（2回目以降の全体実行は約5秒） |
| 認証 | ミドルウェアは存在したがルートに未適用（API は常に 401） | Dating / M&A の全ルートに JWT 認証を適用（`internal/shared/auth` に検証を共通化） |
| 配線 | `cmd/api/main.go` に直書き | `internal/app` に集約し、E2E テストと共用 |
| E2E テスト | なし | `test/e2e` で正常・異常の主要導線を検証 |
| 品質ゲート | gofmt / vet / staticcheck / golangci-lint / カバレッジ 70% | 上記に加え govulncheck、`sqlc diff`、コメント形式検査、カバレッジ 80%、CRAP < 15、mutation testing（gremlins、効力 80% 以上） |
| M&A 設定 | 旧スキーマで起動時に検証エラー | 現行スキーマに書き直し |
| M&A 関心表明 ID | 定数（2件目で主キー衝突） | 暗号学的乱数 |
| スパース特徴 | タグ・技術・市場が空だと Jaccard がエラーになり候補が全て落ちる | 空集合としてキーを確保 |
| プロフィール読込 | タグ・写真を読み込まない | タグ・写真も読み込む |
| DSN | パスワード中の `@` などが未エスケープ | `url.URL` で組み立て |

## 技術スタック

| カテゴリ | 選択 |
| ------- | ---- |
| 言語 | Go 1.25+ |
| DB | PostgreSQL 18 |
| データアクセス | sqlc + pgx/v5 |
| テスト | testing、testcontainers-go（PostgreSQL 18、コンテナ再利用）、httptest |
| 品質 | gofmt、go vet、staticcheck、golangci-lint、govulncheck、gremlins（mutation testing）、tools/crap（CRAP 値） |
| コンテナ | Docker + docker compose |

## クイックスタート

```bash
# 開発環境起動（app + PostgreSQL 18）
docker compose up -d

# テスト実行（Docker が必要。初回は PostgreSQL 18 イメージを取得する）
go test -shuffle=on -count=1 ./...

# 品質チェック一括（CI と同じゲート）
./scripts/check.sh
# mutation testing を省略して高速確認
SKIP_MUTATION=1 ./scripts/check.sh

# ビルド
go build -o matching-engine ./cmd/api
```

### colima を使っている場合

testcontainers が Docker ソケットを見つけられるように環境変数を設定します。

```bash
export DOCKER_HOST=unix://$HOME/.colima/default/docker.sock
export TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE=/var/run/docker.sock
```

## テストデータベースの仕組み

`internal/testutil` が次の順で PostgreSQL を用意します。

1. `matching-engine-pg18-test` という名前のコンテナがあれば再利用し、無ければ起動します（`WithReuseByName`）。再利用コンテナは Ryuk に回収させないため、プロセス終了後も残ります。
2. マイグレーション済みのテンプレート DB `matching_template` を一度だけ作ります。マイグレーションファイルのハッシュを DB に記録し、変更があれば作り直します。
3. テストプロセスごとに `CREATE DATABASE test_<時刻>_<乱数> TEMPLATE matching_template` で独立した DB を払い出します。パッケージ間で干渉しません。
4. 起動時に、30分より古く接続の無い `test_*` DB を削除します。

環境変数で挙動を変えられます。

| 変数 | 意味 |
| ---- | ---- |
| `TEST_DATABASE_ADMIN_DSN` | 既存の PostgreSQL を使う（CI のサービスコンテナなど）。管理 DB への接続文字列 |
| `TESTCONTAINERS_REUSE=false` | コンテナ再利用を無効にし、テストごとに起動・破棄する（CI 既定） |

コンテナを片付けるには `docker rm -f matching-engine-pg18-test` を実行します。

## 品質ゲート

`./scripts/check.sh` と CI は同じゲートを通します。

| ゲート | 内容 |
| ------ | ---- |
| gofmt / go vet / staticcheck / golangci-lint | 静的検査（設定は `.golangci.yml`。生成コードは対象外） |
| govulncheck | 既知の脆弱性 |
| sqlc diff | 生成コードが `queries/*.sql` と一致しているか |
| コメント形式 | `// Name はXXX` 形式を禁止し `// Name XXX` に統一 |
| go test -race | 全テスト（testcontainers で PostgreSQL 18） |
| カバレッジ | 80% 以上（sqlcgen / cmd / tools を除く） |
| CRAP | 全関数で 15 未満（`go run ./tools/crap -profile coverage.out -threshold 15`） |
| mutation testing | gremlins。DB 非依存パッケージが対象で効力 80% 以上（設定は `.gremlins.yaml`） |

mutation testing の対象からリポジトリ層・E2E・配線・生成コードを外しているのは、変異体ごとに testcontainers を起動すると実行時間が非現実的になるためです。それらの正しさは通常のテストと E2E で担保します。

## sqlc

クエリは `internal/modules/<module>/infrastructure/repository/queries/*.sql`、設定は `sqlc.yaml`、生成先は各モジュールの `repository/sqlcgen` です。

```bash
sqlc generate   # 生成
sqlc diff       # 生成漏れの検出
```

可変条件の検索は `sqlc.narg()` と NULL 判定で1本のクエリにまとめています（`FindCandidateUsers`、`ListCompaniesByPurpose`）。

## ディレクトリ構造

```text
matching-engine/
├── cmd/api/                 # エントリーポイント
├── internal/
│   ├── app/                 # 配線（ルーター構築）
│   ├── core/matching/       # コアエンジン
│   ├── modules/
│   │   ├── dating/          # Datingモジュール（api / application / domain / infrastructure）
│   │   └── ma/              # M&Aモジュール
│   ├── shared/              # 共有インフラ（auth / config / database / health / logger）
│   └── testutil/            # テスト用 PostgreSQL
├── db/migrations/           # スキーマ（PostgreSQL 18）
├── db/testdata/             # シードデータ
├── configs/                 # マッチング設定（JSON）
├── test/e2e/                # E2E テスト
├── tools/crap/              # CRAP 計測ツール
└── scripts/                 # ユーティリティ
```

## API

全ルートは `Authorization: Bearer <JWT>`（HS256、秘密鍵は `JWT_SECRET`）を要求します。Dating は `user_id` クレーム、M&A は `company_id` クレームです。

| メソッド | パス | 内容 |
| ------ | ---- | ---- |
| GET | /health/live, /health/ready | ヘルスチェック（認証不要） |
| GET | /api/v1/dating/matches?limit= | 候補取得 |
| POST | /api/v1/dating/likes | いいね送信 `{"target_user_id": "..."}` |
| GET | /api/v1/dating/likes/received | 受け取ったいいね |
| GET | /api/v1/dating/matches/mutual | 相互マッチ |
| GET | /api/v1/ma/targets?limit= | 候補企業 |
| POST | /api/v1/ma/interests | 関心表明 `{"target_company_id": "..."}` |
| GET | /api/v1/ma/interests/received | 受け取った関心表明 |
| GET | /api/v1/ma/matches | 相互マッチ |
| GET | /api/v1/ma/valuation/{id} | バリュエーション（未実装、501） |

## 開発

```bash
go fmt ./...
go vet ./...
staticcheck ./...
golangci-lint run ./...
govulncheck ./...
sqlc diff
go test -shuffle=on -count=1 -race ./...
gremlins unleash
```

コードコメントは「`// Name 説明`」の形式で書きます（「`// Name は説明`」は CI で拒否します）。

## ライセンス

MIT
