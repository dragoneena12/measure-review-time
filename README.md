# Measure Review Time

GitHub PRのオープンからApproveまでの時間を計測するCLIツール

## セットアップ

```bash
# 依存関係のダウンロード
go mod download
```

## 使用方法

### 基本的な使い方

```bash
# 環境変数でトークンを設定（必須）
export GITHUB_TOKEN=your_github_token

# リポジトリのPRレビュー時間を計測
go run cmd/measure/main.go -owner owner_name -repo repo_name
```

### オプション

- `-owner, -o`: リポジトリオーナー（必須）
- `-repo, -r`: リポジトリ名（必須）
- `-since`: この日付以降のPRのみ分析 (YYYY-MM-DD)
- `-format, -f`: 出力形式 (table, json, csv) デフォルト: table
- `-debug`: デバッグログを有効化

### 環境変数

以下の環境変数の設定が必要です：

- `GITHUB_TOKEN`: GitHub Personal Access Token（必須）

### 例

```bash
# 最近のクローズドPRを分析（最大100件）
go run cmd/measure/main.go -o facebook -r react

# 2024年以降のPRをCSV形式で出力
go run cmd/measure/main.go -o facebook -r react -since 2024-01-01 -f csv

# JSON形式で出力
go run cmd/measure/main.go -o facebook -r react -f json

# デバッグログを有効化
go run cmd/measure/main.go -o facebook -r react -debug
```

## 出力形式

### Table形式（デフォルト）
```
=== PR Review Time Report for facebook/react ===

PR #  Author      Created              Time to Review  Time to Approve  Title
----  ------      -------              --------------  ---------------  -----
12345  john_doe   2024-01-15 10:30    1d 2h           2d 5h            Fix memory leak in useEffect
12344  jane_smith 2024-01-14 14:20    3h 45m          1d 3h            Add new feature for concurrent rendering
```

### CSV形式
```csv
PR_Number,Title,Author,Created_At,Time_To_Review,Time_To_Approve
12345,"Fix memory leak in useEffect",john_doe,2024-01-15 10:30:00,1d 2h,2d 5h
12344,"Add new feature for concurrent rendering",jane_smith,2024-01-14 14:20:00,3h 45m,1d 3h
```

### JSON形式
```json
{
  "repository": "facebook/react",
  "pull_requests": [
    {
      "number": 12345,
      "title": "Fix memory leak in useEffect",
      "author": "john_doe",
      "created_at": "2024-01-15T10:30:00Z",
      "time_to_review": "1d 2h",
      "time_to_approve": "2d 5h"
    },
    {
      "number": 12344,
      "title": "Add new feature for concurrent rendering",
      "author": "jane_smith",
      "created_at": "2024-01-14T14:20:00Z",
      "time_to_review": "3h 45m",
      "time_to_approve": "1d 3h"
    }
  ]
}
```

## 計測される指標

- **Time to Review**: レビューリクエストから最初の人間によるレビューまでの時間（レビューリクエストがない場合はPR作成時刻から）
- **Time to Approve**: レビューリクエストから最初のApproveまでの時間（レビューリクエストがない場合はPR作成時刻から）

注意事項：
- GitHub Apps（bot）からのレビューは除外されます
- レビューリクエスト前のレビューは計測対象外です

## アーキテクチャ

Clean Architectureに基づいた3層構造：

- **Domain Layer**: ビジネスロジックとエンティティ
- **Infrastructure Layer**: GitHub APIとの通信
- **Application Layer**: ユースケース実装

## 必要な権限

GitHub Personal Access Tokenには以下の権限が必要です：
- `repo` (プライベートリポジトリの場合)
- `public_repo` (パブリックリポジトリのみの場合)

## トークンの作成方法

1. GitHubにログイン
2. Settings → Developer settings → Personal access tokens → Tokens (classic)
3. "Generate new token"をクリック
4. 必要なスコープを選択（repo または public_repo）
5. トークンを生成してコピー

## 注意事項

- GitHub APIのレート制限に注意してください（認証済み: 5000リクエスト/時）
- 大量のPRを分析する場合は、`--limit`オプションで制限することを推奨
- プライベートリポジトリの場合は適切な権限を持つトークンが必要です