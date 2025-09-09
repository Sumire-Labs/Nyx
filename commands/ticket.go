package commands

import (
	"fmt"
	"time"

	nyxcomponents "github.com/Sumire-Labs/Nyx-API/components"
	nyxembed "github.com/Sumire-Labs/Nyx-API/embed"
	"github.com/Sumire-Labs/Nyx/database"
	"github.com/Sumire-Labs/Nyx/utils"
	"github.com/bwmarrin/discordgo"
)

func init() {
	ticketCommand := &Command{
		Name:                "ticket",
		Description:         "チケットシステムの管理",
		Usage:               "ticket create [title] [description]",
		Aliases:             []string{"t"},
		Category:            "Admin",
		GuildOnly:           true,
		OwnerOnly:           false,
		RequiredPermissions: discordgo.PermissionManageChannels,
		SlashCommand: &discordgo.ApplicationCommand{
			Name:        "ticket",
			Description: "チケットシステムの管理",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "create",
					Description: "チケット作成パネルを設置します",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "title",
							Description: "パネルのタイトル",
							Required:    false,
						},
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "description",
							Description: "パネルの説明",
							Required:    false,
						},
					},
				},
			},
		},
		Execute:      executeTicket,
		ExecuteSlash: executeTicketSlash,
	}

	DefaultCommands = append(DefaultCommands, ticketCommand)
}

func executeTicket(ctx *Context) error {
	if len(ctx.Args) == 0 {
		return ctx.ReplyError("使用方法: `ticket create [title] [description]`")
	}

	switch ctx.Args[0] {
	case "create":
		return executeTicketCreate(ctx)
	default:
		return ctx.ReplyError("無効なサブコマンドです。使用可能: `create`")
	}
}

func executeTicketSlash(ctx *SlashContext) error {
	subcommand := ctx.Interaction.ApplicationCommandData().Options[0].Name
	
	switch subcommand {
	case "create":
		return executeTicketCreateSlash(ctx)
	default:
		return ctx.ReplyError("無効なサブコマンドです", true)
	}
}

func executeTicketCreate(ctx *Context) error {
	title := "🎫 サポートチケット"
	description := "下のボタンをクリックしてサポートチケットを作成してください。"

	if len(ctx.Args) > 1 {
		// 🔧 FIXED: タイトル入力検証を追加
		if err := utils.ValidateMessageContent(ctx.Args[1]); err != nil {
			return ctx.ReplyError("無効なタイトルです: " + err.Error())
		}
		title = utils.SanitizeInput(ctx.Args[1])
	}
	if len(ctx.Args) > 2 {
		// 🔧 FIXED: 説明入力検証を追加
		if err := utils.ValidateMessageContent(ctx.Args[2]); err != nil {
			return ctx.ReplyError("無効な説明です: " + err.Error())
		}
		description = utils.SanitizeInput(ctx.Args[2])
	}

	return createTicketPanel(ctx.Session, ctx.Message.ChannelID, ctx.Message.GuildID, ctx.DB, title, description)
}

func executeTicketCreateSlash(ctx *SlashContext) error {
	title := "🎫 サポートチケット"
	description := "下のボタンをクリックしてサポートチケットを作成してください。"

	options := ctx.Interaction.ApplicationCommandData().Options[0].Options
	for _, option := range options {
		switch option.Name {
		case "title":
			// 🔧 FIXED: スラッシュコマンドでもタイトル検証追加
			if err := utils.ValidateMessageContent(option.StringValue()); err != nil {
				return ctx.ReplyError("無効なタイトルです: "+err.Error(), true)
			}
			title = utils.SanitizeInput(option.StringValue())
		case "description":
			// 🔧 FIXED: スラッシュコマンドでも説明検証追加
			if err := utils.ValidateMessageContent(option.StringValue()); err != nil {
				return ctx.ReplyError("無効な説明です: "+err.Error(), true)
			}
			description = utils.SanitizeInput(option.StringValue())
		}
	}

	err := createTicketPanel(ctx.Session, ctx.Interaction.ChannelID, ctx.Interaction.GuildID, ctx.DB, title, description)
	if err != nil {
		return ctx.ReplyError("パネルの作成に失敗しました: "+err.Error(), true)
	}

	return ctx.ReplySuccess("チケットパネルを作成しました！", true)
}

func createTicketPanel(session *discordgo.Session, channelID, guildID string, db *database.Database, title, description string) error {
	embed := nyxembed.Info(title, description).
		SetColor(0x5865F2).
		SetFooter(nyxembed.Footer{
			Text: "サポートが必要な場合は、下のボタンをクリックしてください",
		}).
		SetTimestamp(time.Now()).
		Build()

	button := nyxcomponents.PrimaryButton("🎫 チケット作成", "create_ticket")
	actionRow := nyxcomponents.CreateButtonRow(button)

	message, err := session.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{
		Embeds:     []*discordgo.MessageEmbed{embed},
		Components: []discordgo.MessageComponent{actionRow},
	})
	if err != nil {
		return fmt.Errorf("メッセージ送信エラー: %w", err)
	}

	panel := &database.TicketPanel{
		GuildID:     guildID,
		ChannelID:   channelID,
		MessageID:   message.ID,
		Title:       title,
		Description: &description,
	}

	if err := db.CreateTicketPanel(panel); err != nil {
		return fmt.Errorf("データベースエラー: %w", err)
	}

	return nil
}