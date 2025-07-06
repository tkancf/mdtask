# mdtask

- mdtaskはMarkdownファイルをタスク管理のチケットとして扱うためのツール

## Markdownファイル フォーマット

タスク管理のチケットとして扱うMarkdownファイルはYAML frontmatterを持ち、tagsに mdtask タグを持つ

```yaml
---
id: unique-identifier
aliases:
  - Alternative Title
tags:
    - mdtask
created: YYYY-MM-DD HH:MM
description: Brief description
title: Display Title
updated: YYYY-MM-DD HH:MM
---

# Content Here
```

- unique-identifier = task/YYYYMMDDHHMMSS
- YYYYMMDDHHMMSS はファイル作成日時分秒

### タスクの管理

タスクのステータス、管理対象か否かの判断には、YAML frontmatterのtagsを利用する

- mdtaskタグを持つ場合
    - mdtaskの管理対象
- タスクのステータスは `mdtsk/status/***` で管理する
    - TODOステータス: mdtask/status/TODO
    - 進行中ステータス: mdtask/status/WIP
    - 相手ボールの対応を待っているステータス: mdtask/status/WAIT
    - スケジュール済みでその時間を待っているステータス: mdtask/status/SCHE
    - 完了ステータス: mdtask/status/DONE
- タスクのアーカイブは `mdtask/archived` で管理する
    - アーカイブ済みタスク: mdtask/archived
- タスクの期日は `mdtask/deadline/YYYY-MM-DD` で管理する
    - 2025/06/29期日のタスク: mdtask/deadline/2025-06-29
- 待ちステータス(`mdtask/status/WAIT`)の理由は `mdtask/waitfor/****` で管理する
    - 相手のメール返信待ちのタスク `mdtask/waitfor/メール返信待ち`

## mdtaskの機能

- 上記フォーマットのMarkdownファイルを管理、作成できる
- Go言語で実装
- mdtaskはCLIインターフェイスを提供する
    - `mdtask list` - タスクの一覧（--status, --archived, --allオプション付き）
    - `mdtask search [query]` - タスクの検索
    - `mdtask new` - タスクの作成（対話的またはフラグ指定）
    - `mdtask edit [task-id]` - タスクの編集（エディタ起動）
    - `mdtask archive [task-id]` - タスクのアーカイブ
    - `mdtask tui` - ターミナルUIの起動（インタラクティブなタスク管理）
- mdtaskはWebブラウザインターフェイスを提供する
    - `mdtask web` - WebUIの起動（デフォルトポート: 7000、自動ポート切替機能付き）
    - ダッシュボード、タスク管理、検索機能を含む直感的なUI
- mdtaskはMCP (Model Context Protocol) サーバーを提供する
    - `mdtask mcp` - MCPサーバーの起動（AI assistants向け）
    - Claude DesktopなどのMCP対応ツールからタスクを管理可能
- mdtaskの設定
    - TOML形式の設定ファイルをサポート（.mdtask.toml、mdtask.toml、~/.config/mdtask/config.toml、~/.mdtask.toml）
    - 設定可能な項目：
        - `paths` - 管理対象ディレクトリの指定
        - `task.title_prefix` - タスクタイトルに自動付与するプレフィックス
        - `task.default_status` - 新規タスクのデフォルトステータス
        - `web.port` - WebUIのデフォルトポート番号
        - `web.open_browser` - WebUI起動時のブラウザ自動起動設定
        - `mcp.enabled` - MCPサーバーの有効/無効
        - `mcp.allowed_paths` - MCPサーバーがアクセス可能な追加パス

## インストール

### 前提条件

- Go 1.19以上
- Node.js 16以上（WebUIのスタイルとJavaScript生成用）

### ソースからビルド

```bash
git clone https://github.com/tkancf/mdtask.git
cd mdtask

# 依存関係のインストールとビルド
make

# または個別に実行
npm install
npm run build
go build -o mdtask
```

### Makefileターゲット

- `make` - 依存関係のインストール、CSS/JavaScript生成、バイナリビルド
- `make build` - バイナリのビルド（CSS/JavaScript生成含む）
- `make css` - CSSのみビルド
- `make js` - JavaScriptのみビルド（TypeScriptのコンパイル）
- `make watch` - CSS変更の監視（開発用）
- `make test` - テストの実行
- `make release` - 全プラットフォーム向けリリースビルド
- `make clean` - ビルド成果物のクリーン
- `make install` - ローカルインストール（/usr/local/bin）

### 開発モード

開発中はCSS/JavaScriptの変更を監視できます：

```bash
# CSSの変更を監視
npm run watch-css

# 別のターミナルでTypeScript/JavaScriptの変更を監視
npm run dev-js

# さらに別のターミナルでGoアプリケーションを実行
go run main.go web
```

### 技術スタック

- **バックエンド**: Go 1.19+
  - Cobra (CLI framework)
  - Chi (HTTP router)
  - mark3labs/mcp-go (MCP implementation)
- **フロントエンド**: 
  - TypeScript (型安全なJavaScript)
  - Vite (高速なビルドツール)
  - Tailwind CSS (ユーティリティファーストCSS)
- **ビルドツール**: 
  - Make (ビルド自動化)
  - npm (パッケージ管理)

## MCP (Model Context Protocol) サーバー

mdtaskはMCPサーバーを内蔵しており、Claude DesktopなどのMCP対応AIアシスタントからタスクを管理できます。

### MCP設定

Claude Desktopで使用する場合は、`claude_desktop_config.json`に以下を追加：

```json
{
  "mcpServers": {
    "mdtask": {
      "command": "/path/to/mdtask",
      "args": ["mcp"],
      "cwd": "/path/to/your/tasks"
    }
  }
}
```

### 利用可能なMCPツール

- `list_tasks` - タスクの一覧表示（ステータスフィルタ、アーカイブ表示対応）
- `create_task` - 新規タスクの作成
- `update_task` - タスクの更新（タイトル、説明、ステータス、タグ）
- `search_tasks` - タスクの検索
- `archive_task` - タスクのアーカイブ
- `get_task` - 特定タスクの詳細取得
- `get_statistics` - タスク統計の取得

### 利用可能なMCPリソース

- `tasks` - アクティブなタスクのMarkdown形式リスト
- `statistics` - タスク統計情報（JSON形式）

## Neovimプラグイン

mdtaskには、Neovimから直接タスクを管理できるプラグインが含まれています。

### インストール

プラグインは`nvim-mdtask`サブディレクトリにあります。

**lazy.nvimの場合:**
```lua
{
  dir = '~/path/to/mdtask/nvim-mdtask',  -- mdtaskリポジトリのパスを指定
  name = 'mdtask.nvim',
  dependencies = {
    'nvim-telescope/telescope.nvim', -- optional
  },
  config = function()
    require('mdtask').setup()
  end,
}
```

### 主な機能

- `:MdTask` - タスク一覧表示
- `:MdTask new` - 新規タスク作成
- `:MdTask search <query>` - タスク検索
- `:MdTask status <status>` - ステータス別表示

詳細は[nvim-mdtask/README.md](nvim-mdtask/README.md)を参照してください。

## 開発

### アーキテクチャ

mdtaskは以下の層で構成されています：

- **CLIコマンド層** (`cmd/mdtask/`): ユーザーインターフェース
- **サービス層** (`internal/service/`): ビジネスロジック
- **リポジトリ層** (`internal/repository/`): データアクセス
- **共通ユーティリティ** (`internal/cli/`, `internal/output/`): 横断的関心事

### ビルドとテスト

```bash
# ビルド
go build -o mdtask

# テスト実行
go test ./...

# テストとリント実行
./test.sh
```

### コード構造

```
mdtask/
├── cmd/mdtask/          # CLIコマンド
│   ├── root.go         # ルートコマンド
│   ├── new.go          # タスク作成
│   ├── list.go         # タスク一覧
│   └── ...
├── internal/           # 内部パッケージ
│   ├── cli/           # CLI共通ユーティリティ
│   ├── service/       # ビジネスロジック層
│   ├── repository/    # データアクセス層
│   ├── task/          # タスクモデル
│   └── config/        # 設定管理
├── pkg/               # 公開パッケージ
│   └── markdown/      # Markdownパーサー
└── nvim-mdtask/       # Neovimプラグイン
```

### 貢献

1. このリポジトリをフォーク
2. フィーチャーブランチを作成 (`git checkout -b feature/amazing-feature`)
3. 変更をコミット (`git commit -m 'Add some amazing feature'`)
4. ブランチにプッシュ (`git push origin feature/amazing-feature`)
5. プルリクエストを作成
