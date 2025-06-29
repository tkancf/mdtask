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
- mdtaskはWebブラウザインターフェイスを提供する
    - `mdtask web` - WebUIの起動（デフォルトポート: 7000、自動ポート切替機能付き）
    - ダッシュボード、タスク管理、検索機能を含む直感的なUI
- mdtaskの設定
    - TOML形式の設定ファイルをサポート（.mdtask.toml、mdtask.toml、~/.config/mdtask/config.toml、~/.mdtask.toml）
    - 設定可能な項目：
        - `paths` - 管理対象ディレクトリの指定
        - `task.title_prefix` - タスクタイトルに自動付与するプレフィックス
        - `task.default_status` - 新規タスクのデフォルトステータス
        - `web.port` - WebUIのデフォルトポート番号
        - `web.open_browser` - WebUI起動時のブラウザ自動起動設定

## インストール

### 前提条件

- Go 1.19以上
- Node.js 16以上（WebUIのスタイル生成用）

### ソースからビルド

```bash
git clone https://github.com/tkan/mdtask.git
cd mdtask

# 依存関係のインストールとビルド
make

# または個別に実行
npm install
npm run build-css
go build -o mdtask
```

### Makefileターゲット

- `make` - 依存関係のインストール、CSS生成、バイナリビルド
- `make build` - バイナリのビルド（CSS生成含む）
- `make css` - CSSのみビルド
- `make watch` - CSS変更の監視（開発用）
- `make test` - テストの実行
- `make release` - 全プラットフォーム向けリリースビルド
- `make clean` - ビルド成果物のクリーン
- `make install` - ローカルインストール（/usr/local/bin）

### 開発モード

開発中はCSSの変更を監視できます：

```bash
# CSSの変更を監視
npm run watch-css

# 別のターミナルでGoアプリケーションを実行
go run main.go web
```
