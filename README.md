# Nyx 1.0.0

**Go言語で構築されたALL-IN-ONE DISCORDBOT**

Nyxは、Discord.goライブラリを基盤とし、独自開発のNyx-APIを活用して作られた高性能なDiscord Botです。モジュラー設計により拡張性に優れ、SQLiteによるデータ永続化と豊富なコマンド機能を提供します。

## 現在の機能

- **チケットツール** - プライベートなサポートチケット作成・管理
- **ユーザー情報表示** - アバター・バナー・プロフィール情報の表示
- **レイテンシ計測** - Ping・レイテンシ・ヘルスチェック
- **データベース統合** - SQLiteによる永続的なデータ保存
- **モジュラー設計** - 簡単なコマンド追加・拡張

## プロジェクト構造

```
Nyx/
├── main.go              # エントリーポイント
├── config.yaml          # 設定ファイル
├── bot/                 # Bot コア機能
│   ├── bot.go          # メインBot構造体
│   └── ticket_handler.go # チケット機能ハンドラー
├── commands/           # コマンド実装
│   ├── registry.go     # コマンド登録システム
│   ├── ping.go         # Pingコマンド
│   ├── avatar.go       # Avatarコマンド
│   └── ticket.go       # チケットコマンド
├── database/           # データベース層
│   ├── database.go     # DB接続・マイグレーション
│   └── models.go       # データモデル・CRUD操作
└── logs/               # ログファイル
    └── bot.log
```

## 設定オプション

### Bot設定
```yaml
bot:
  token: "string"        # Bot Token (必須)
  prefix: "string"       # コマンドプレフィックス (デフォルト: "!")
  status: "string"       # Bot ステータス (online/idle/dnd/invisible)
  activity:
    type: "string"       # アクティビティタイプ
    name: "string"       # アクティビティ名
```

### 機能設定
```yaml
features:
  slash_commands: bool   # スラッシュコマンド有効化
  logging_channel: "id"  # ログ送信チャンネル
```

### ログ設定
```yaml
logging:
  level: "string"        # ログレベル (debug/info/warn/error)
  file: "path"          # ログファイルパス
  webhook_url: "url"    # Discord Webhook URL (エラー通知用)
```

## 開発

### 新しいコマンドの追加

1. **commands/ ディレクトリに新しいファイルを作成**
```go
// commands/example.go
package commands

func init() {
    exampleCommand := &Command{
        Name:        "example",
        Description: "例コマンド",
        Execute:     executeExample,
        ExecuteSlash: executeExampleSlash,
    }
    DefaultCommands = append(DefaultCommands, exampleCommand)
}

func executeExample(ctx *Context) error {
    return ctx.Reply("Hello, World!")
}
```

2. **Bot再起動で自動的にロード**

### データベースモデルの追加

1. **database/models.go にモデルを追加**
2. **database/database.go にマイグレーション追加**
3. **CRUD操作メソッドを実装**

## 技術スタック

- **言語**: Go 1.24+
- **Discordライブラリ**: [DiscordGo](https://github.com/bwmarrin/discordgo)
- **データベース**: SQLite3
- **設定**: YAML
- **内部API**: [Nyx-API v0.2.0](https://github.com/Sumire-Labs/Nyx-API)

## データベーススキーマ

### テーブル一覧
- `guilds` - Discord サーバー情報
- `users` - ユーザー情報
- `user_guilds` - ユーザー・サーバー関係
- `command_logs` - コマンド実行ログ
- `tickets` - チケット情報
- `ticket_panels` - チケットパネル情報

## ライセンス

このプロジェクトはMPL-2.0ライセンスの下で公開されています。詳細は [LICENSE.md](LICENSE.md) を参照してください。

## 関連リンク

- [Nyx-API](https://github.com/Sumire-Labs/Nyx-API) - 内部API・コンポーネントライブラリ
- [DiscordGo](https://github.com/bwmarrin/discordgo) - Discord API ラッパー