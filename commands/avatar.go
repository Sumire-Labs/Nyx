package commands

import (
	"fmt"

	nyxembed "github.com/Sumire-Labs/Nyx-API/embed"
	"github.com/Sumire-Labs/Nyx/utils"
	"github.com/bwmarrin/discordgo"
)

func init() {
	avatarCommand := &Command{
		Name:        "avatar",
		Description: "ユーザーのアバターとバナーを表示します",
		Usage:       "avatar [@user]",
		Aliases:     []string{"av", "pfp"},
		Category:    "Utility",
		GuildOnly:   false,
		OwnerOnly:   false,
		SlashCommand: &discordgo.ApplicationCommand{
			Name:        "avatar",
			Description: "ユーザーのアバターとバナーを表示します",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionUser,
					Name:        "user",
					Description: "アバターを表示するユーザー",
					Required:    false,
				},
			},
		},
		Execute:      executeAvatar,
		ExecuteSlash: executeAvatarSlash,
	}

	DefaultCommands = append(DefaultCommands, avatarCommand)
}

func executeAvatar(ctx *Context) error {
	var targetUser *discordgo.User
	
	if len(ctx.Args) > 0 {
		// 🔧 FIXED: 安全な入力検証を追加
		userID := utils.ExtractUserIDFromMention(ctx.Args[0])
		if userID == "" {
			return ctx.ReplyError("無効なユーザーメンション形式です。正しい形式: @user")
		}
		
		if !utils.ValidateDiscordID(userID) {
			return ctx.ReplyError("無効なユーザーIDです。")
		}
		
		user, err := ctx.Session.User(userID)
		if err != nil {
			return ctx.ReplyError("指定されたユーザーが見つかりませんでした。")
		}
		targetUser = user
	} else {
		targetUser = ctx.Message.Author
	}

	return sendAvatarEmbed(ctx.Session, ctx.Message.ChannelID, targetUser)
}

func executeAvatarSlash(ctx *SlashContext) error {
	var targetUser *discordgo.User
	
	userOption := ctx.GetOptionValue("user")
	if userOption != nil {
		targetUser = userOption.UserValue(ctx.Session)
	} else {
		targetUser = ctx.Interaction.Member.User
	}

	embed := createAvatarEmbed(targetUser)
	return ctx.ReplyEmbed(embed, false)
}

func sendAvatarEmbed(session *discordgo.Session, channelID string, user *discordgo.User) error {
	embed := createAvatarEmbed(user)
	_, err := session.ChannelMessageSendEmbed(channelID, embed)
	return err
}

func createAvatarEmbed(user *discordgo.User) *discordgo.MessageEmbed {
	embed := nyxembed.Info("", "").
		SetTitle(fmt.Sprintf("%s のアバター", user.Username)).
		SetColor(0x5865F2)

	// アバターURL（サイズ指定）
	avatarURL := user.AvatarURL("2048")
	if avatarURL != "" {
		embed.SetImage(nyxembed.Image{URL: avatarURL}).
			AddField(nyxembed.Field{
				Name:   "🖼️ アバター",
				Value:  fmt.Sprintf("[PNG](%s) | [JPG](%s) | [WEBP](%s)", 
					user.AvatarURL("2048") + "?format=png",
					user.AvatarURL("2048") + "?format=jpg", 
					user.AvatarURL("2048") + "?format=webp"),
				Inline: false,
			})
	} else {
		embed.SetDescription("このユーザーはカスタムアバターを設定していません。")
	}

	// バナー取得を試行（フルユーザーオブジェクトが必要）
	if user.Banner != "" {
		bannerURL := discordgo.EndpointUserBanner(user.ID, user.Banner) + "?size=2048"
		embed.AddField(nyxembed.Field{
			Name:   "🎨 バナー",
			Value:  fmt.Sprintf("[PNG](%s) | [JPG](%s) | [WEBP](%s)",
				bannerURL + "&format=png",
				bannerURL + "&format=jpg",
				bannerURL + "&format=webp"),
			Inline: false,
		})
	}

	// ユーザー情報
	embed.AddField(nyxembed.Field{
		Name:   "👤 ユーザー情報",
		Value:  fmt.Sprintf("**名前:** %s\n**ID:** %s", user.Username, user.ID),
		Inline: true,
	})

	// アカウント作成日
	timestamp, _ := discordgo.SnowflakeTimestamp(user.ID)
	embed.AddField(nyxembed.Field{
		Name:   "📅 アカウント作成日",
		Value:  timestamp.Format("2006/01/02 15:04:05"),
		Inline: true,
	})

	return embed.Build()
}