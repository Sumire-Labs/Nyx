package commands

import (
	"fmt"
	"strings"

	nyxembed "github.com/Sumire-Labs/Nyx-API/embed"
	"github.com/Sumire-Labs/Nyx/database"
	"github.com/bwmarrin/discordgo"
)

func init() {
	loggerCommand := &Command{
		Name:                "logger",
		Description:         "サーバーログの設定・管理",
		Usage:               "logger [enable|disable|setup|status]",
		Aliases:             []string{"log", "logging"},
		Category:            "Admin",
		GuildOnly:           true,
		OwnerOnly:           false,
		RequiredPermissions: discordgo.PermissionManageGuild,
		SlashCommand: &discordgo.ApplicationCommand{
			Name:        "logger",
			Description: "サーバーログの設定・管理",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "enable",
					Description: "ログ機能を有効にします",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
				},
				{
					Name:        "disable",
					Description: "ログ機能を無効にします",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
				},
				{
					Name:        "setup",
					Description: "ログチャンネルを設定します",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionChannel,
							Name:        "channel",
							Description: "ログを送信するチャンネル",
							Required:    true,
							ChannelTypes: []discordgo.ChannelType{
								discordgo.ChannelTypeGuildText,
							},
						},
					},
				},
				{
					Name:        "status",
					Description: "現在のログ設定を表示します",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
				},
				{
					Name:        "configure",
					Description: "詳細なログ設定を行います",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "type",
							Description: "設定するログタイプ",
							Required:    true,
							Choices: []*discordgo.ApplicationCommandOptionChoice{
								{Name: "メンバー参加", Value: "member_join"},
								{Name: "メンバー退出", Value: "member_leave"},
								{Name: "メッセージ編集", Value: "message_edit"},
								{Name: "メッセージ削除", Value: "message_delete"},
								{Name: "ロール作成", Value: "role_create"},
								{Name: "ロール更新", Value: "role_update"},
								{Name: "ロール削除", Value: "role_delete"},
								{Name: "ニックネーム変更", Value: "nickname_change"},
								{Name: "BAN", Value: "ban"},
								{Name: "BAN解除", Value: "unban"},
								{Name: "KICK", Value: "kick"},
								{Name: "タイムアウト", Value: "timeout"},
								{Name: "チャンネル作成", Value: "channel_create"},
								{Name: "チャンネル更新", Value: "channel_update"},
								{Name: "チャンネル削除", Value: "channel_delete"},
							},
						},
						{
							Type:        discordgo.ApplicationCommandOptionBoolean,
							Name:        "enabled",
							Description: "有効/無効",
							Required:    true,
						},
					},
				},
			},
		},
		Execute:      executeLogger,
		ExecuteSlash: executeLoggerSlash,
	}

	DefaultCommands = append(DefaultCommands, loggerCommand)
}

func executeLogger(ctx *Context) error {
	if len(ctx.Args) == 0 {
		return showLoggerStatus(ctx)
	}

	switch ctx.Args[0] {
	case "enable":
		return enableLogger(ctx)
	case "disable":
		return disableLogger(ctx)
	case "setup":
		return setupLoggerChannel(ctx)
	case "status":
		return showLoggerStatus(ctx)
	default:
		return ctx.ReplyError("使用方法: `logger [enable|disable|setup|status]`")
	}
}

func executeLoggerSlash(ctx *SlashContext) error {
	subcommand := ctx.Interaction.ApplicationCommandData().Options[0].Name

	switch subcommand {
	case "enable":
		return enableLoggerSlash(ctx)
	case "disable":
		return disableLoggerSlash(ctx)
	case "setup":
		return setupLoggerChannelSlash(ctx)
	case "status":
		return showLoggerStatusSlash(ctx)
	case "configure":
		return configureLoggerSlash(ctx)
	default:
		return ctx.ReplyError("無効なサブコマンドです", true)
	}
}

func enableLogger(ctx *Context) error {
	settings, err := getOrCreateLogSettings(ctx.DB, ctx.Message.GuildID)
	if err != nil {
		return ctx.ReplyError("設定の取得に失敗しました")
	}

	settings.Enabled = true
	if err := ctx.DB.CreateOrUpdateLogSettings(settings); err != nil {
		return ctx.ReplyError("設定の保存に失敗しました")
	}

	if settings.LogChannelID == nil {
		return ctx.ReplySuccess("ログ機能が有効になりました！\n`logger setup #チャンネル名` でログチャンネルを設定してください。")
	}

	return ctx.ReplySuccess(fmt.Sprintf("ログ機能が有効になりました！\nログチャンネル: <#%s>", *settings.LogChannelID))
}

func enableLoggerSlash(ctx *SlashContext) error {
	settings, err := getOrCreateLogSettings(ctx.DB, ctx.Interaction.GuildID)
	if err != nil {
		return ctx.ReplyError("設定の取得に失敗しました", true)
	}

	settings.Enabled = true
	if err := ctx.DB.CreateOrUpdateLogSettings(settings); err != nil {
		return ctx.ReplyError("設定の保存に失敗しました", true)
	}

	if settings.LogChannelID == nil {
		return ctx.ReplySuccess("ログ機能が有効になりました！\n`/logger setup` でログチャンネルを設定してください。", false)
	}

	return ctx.ReplySuccess(fmt.Sprintf("ログ機能が有効になりました！\nログチャンネル: <#%s>", *settings.LogChannelID), false)
}

func disableLogger(ctx *Context) error {
	settings, err := getOrCreateLogSettings(ctx.DB, ctx.Message.GuildID)
	if err != nil {
		return ctx.ReplyError("設定の取得に失敗しました")
	}

	settings.Enabled = false
	if err := ctx.DB.CreateOrUpdateLogSettings(settings); err != nil {
		return ctx.ReplyError("設定の保存に失敗しました")
	}

	return ctx.ReplySuccess("ログ機能が無効になりました。")
}

func disableLoggerSlash(ctx *SlashContext) error {
	settings, err := getOrCreateLogSettings(ctx.DB, ctx.Interaction.GuildID)
	if err != nil {
		return ctx.ReplyError("設定の取得に失敗しました", true)
	}

	settings.Enabled = false
	if err := ctx.DB.CreateOrUpdateLogSettings(settings); err != nil {
		return ctx.ReplyError("設定の保存に失敗しました", true)
	}

	return ctx.ReplySuccess("ログ機能が無効になりました。", false)
}

func setupLoggerChannel(ctx *Context) error {
	if len(ctx.Args) < 2 {
		return ctx.ReplyError("使用方法: `logger setup #チャンネル名`")
	}

	channelID := strings.Trim(ctx.Args[1], "<#>")
	channel, err := ctx.Session.Channel(channelID)
	if err != nil {
		return ctx.ReplyError("指定されたチャンネルが見つかりません")
	}

	if channel.GuildID != ctx.Message.GuildID {
		return ctx.ReplyError("このサーバーのチャンネルを指定してください")
	}

	settings, err := getOrCreateLogSettings(ctx.DB, ctx.Message.GuildID)
	if err != nil {
		return ctx.ReplyError("設定の取得に失敗しました")
	}

	settings.LogChannelID = &channelID
	settings.Enabled = true

	if err := ctx.DB.CreateOrUpdateLogSettings(settings); err != nil {
		return ctx.ReplyError("設定の保存に失敗しました")
	}

	return ctx.ReplySuccess(fmt.Sprintf("ログチャンネルを <#%s> に設定しました！", channelID))
}

func setupLoggerChannelSlash(ctx *SlashContext) error {
	options := ctx.Interaction.ApplicationCommandData().Options[0].Options
	channelID := options[0].ChannelValue(ctx.Session).ID

	settings, err := getOrCreateLogSettings(ctx.DB, ctx.Interaction.GuildID)
	if err != nil {
		return ctx.ReplyError("設定の取得に失敗しました", true)
	}

	settings.LogChannelID = &channelID
	settings.Enabled = true

	if err := ctx.DB.CreateOrUpdateLogSettings(settings); err != nil {
		return ctx.ReplyError("設定の保存に失敗しました", true)
	}

	return ctx.ReplySuccess(fmt.Sprintf("ログチャンネルを <#%s> に設定しました！", channelID), false)
}

func configureLoggerSlash(ctx *SlashContext) error {
	options := ctx.Interaction.ApplicationCommandData().Options[0].Options
	logType := options[0].StringValue()
	enabled := options[1].BoolValue()

	settings, err := getOrCreateLogSettings(ctx.DB, ctx.Interaction.GuildID)
	if err != nil {
		return ctx.ReplyError("設定の取得に失敗しました", true)
	}

	updateLogTypeSetting(settings, logType, enabled)

	if err := ctx.DB.CreateOrUpdateLogSettings(settings); err != nil {
		return ctx.ReplyError("設定の保存に失敗しました", true)
	}

	statusText := "無効"
	if enabled {
		statusText = "有効"
	}

	return ctx.ReplySuccess(fmt.Sprintf("ログタイプ「%s」を%sに設定しました", getLogTypeDisplayName(logType), statusText), true)
}

func showLoggerStatus(ctx *Context) error {
	settings, err := getOrCreateLogSettings(ctx.DB, ctx.Message.GuildID)
	if err != nil {
		return ctx.ReplyError("設定の取得に失敗しました")
	}

	embed := buildStatusEmbed(settings)
	return ctx.ReplyEmbed(embed)
}

func showLoggerStatusSlash(ctx *SlashContext) error {
	settings, err := getOrCreateLogSettings(ctx.DB, ctx.Interaction.GuildID)
	if err != nil {
		return ctx.ReplyError("設定の取得に失敗しました", true)
	}

	embed := buildStatusEmbed(settings)
	return ctx.ReplyEmbed(embed, false)
}

func getOrCreateLogSettings(db *database.Database, guildID string) (*database.LogSettings, error) {
	settings, err := db.GetLogSettings(guildID)
	if err != nil {
		return nil, err
	}

	if settings == nil {
		settings = &database.LogSettings{
			GuildID:           guildID,
			Enabled:           true,
			LogMemberJoin:     true,
			LogMemberLeave:    true,
			LogMessageEdit:    true,
			LogMessageDelete:  true,
			LogRoleCreate:     true,
			LogRoleUpdate:     true,
			LogRoleDelete:     true,
			LogNicknameChange: true,
			LogBan:            true,
			LogUnban:          true,
			LogKick:           true,
			LogTimeout:        true,
			LogChannelCreate:  true,
			LogChannelUpdate:  true,
			LogChannelDelete:  true,
		}
	}

	return settings, nil
}

func updateLogTypeSetting(settings *database.LogSettings, logType string, enabled bool) {
	switch logType {
	case "member_join":
		settings.LogMemberJoin = enabled
	case "member_leave":
		settings.LogMemberLeave = enabled
	case "message_edit":
		settings.LogMessageEdit = enabled
	case "message_delete":
		settings.LogMessageDelete = enabled
	case "role_create":
		settings.LogRoleCreate = enabled
	case "role_update":
		settings.LogRoleUpdate = enabled
	case "role_delete":
		settings.LogRoleDelete = enabled
	case "nickname_change":
		settings.LogNicknameChange = enabled
	case "ban":
		settings.LogBan = enabled
	case "unban":
		settings.LogUnban = enabled
	case "kick":
		settings.LogKick = enabled
	case "timeout":
		settings.LogTimeout = enabled
	case "channel_create":
		settings.LogChannelCreate = enabled
	case "channel_update":
		settings.LogChannelUpdate = enabled
	case "channel_delete":
		settings.LogChannelDelete = enabled
	}
}

func getLogTypeDisplayName(logType string) string {
	switch logType {
	case "member_join":
		return "メンバー参加"
	case "member_leave":
		return "メンバー退出"
	case "message_edit":
		return "メッセージ編集"
	case "message_delete":
		return "メッセージ削除"
	case "role_create":
		return "ロール作成"
	case "role_update":
		return "ロール更新"
	case "role_delete":
		return "ロール削除"
	case "nickname_change":
		return "ニックネーム変更"
	case "ban":
		return "BAN"
	case "unban":
		return "BAN解除"
	case "kick":
		return "KICK"
	case "timeout":
		return "タイムアウト"
	case "channel_create":
		return "チャンネル作成"
	case "channel_update":
		return "チャンネル更新"
	case "channel_delete":
		return "チャンネル削除"
	default:
		return logType
	}
}

func buildStatusEmbed(settings *database.LogSettings) *discordgo.MessageEmbed {
	embed := nyxembed.Info("📋 ログ設定状況", "").SetColor(0x5865F2)

	statusIcon := "❌"
	statusText := "無効"
	if settings.Enabled {
		statusIcon = "✅"
		statusText = "有効"
	}

	channelText := "未設定"
	if settings.LogChannelID != nil {
		channelText = fmt.Sprintf("<#%s>", *settings.LogChannelID)
	}

	embed.AddField(nyxembed.Field{
		Name:   "🔧 基本設定",
		Value:  fmt.Sprintf("**ステータス**: %s %s\n**ログチャンネル**: %s", statusIcon, statusText, channelText),
		Inline: false,
	})

	memberEvents := fmt.Sprintf("参加: %s | 退出: %s | ニックネーム: %s",
		boolToIcon(settings.LogMemberJoin),
		boolToIcon(settings.LogMemberLeave),
		boolToIcon(settings.LogNicknameChange))

	messageEvents := fmt.Sprintf("編集: %s | 削除: %s",
		boolToIcon(settings.LogMessageEdit),
		boolToIcon(settings.LogMessageDelete))

	roleEvents := fmt.Sprintf("作成: %s | 更新: %s | 削除: %s",
		boolToIcon(settings.LogRoleCreate),
		boolToIcon(settings.LogRoleUpdate),
		boolToIcon(settings.LogRoleDelete))

	moderationEvents := fmt.Sprintf("BAN: %s | BAN解除: %s | KICK: %s | タイムアウト: %s",
		boolToIcon(settings.LogBan),
		boolToIcon(settings.LogUnban),
		boolToIcon(settings.LogKick),
		boolToIcon(settings.LogTimeout))

	channelEvents := fmt.Sprintf("作成: %s | 更新: %s | 削除: %s",
		boolToIcon(settings.LogChannelCreate),
		boolToIcon(settings.LogChannelUpdate),
		boolToIcon(settings.LogChannelDelete))

	embed.AddField(nyxembed.Field{Name: "👥 メンバーイベント", Value: memberEvents, Inline: false})
	embed.AddField(nyxembed.Field{Name: "💬 メッセージイベント", Value: messageEvents, Inline: false})
	embed.AddField(nyxembed.Field{Name: "🏷️ ロールイベント", Value: roleEvents, Inline: false})
	embed.AddField(nyxembed.Field{Name: "🔨 モデレーション", Value: moderationEvents, Inline: false})
	embed.AddField(nyxembed.Field{Name: "📁 チャンネルイベント", Value: channelEvents, Inline: false})

	return embed.Build()
}

func boolToIcon(b bool) string {
	if b {
		return "✅"
	}
	return "❌"
}