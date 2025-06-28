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
    - mdtaskはtoml形式の設定ファイルで制御が可能
    - 設定には下記を含む
        - 管理対象プロジェクトディレクトリ
            - 指定したディレクトリ配下のMarkdownファイルのみを管理対象とする
