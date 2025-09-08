package bot

import (
	"fmt"
	"strings"
	"time"

	nyxcomponents "github.com/Sumire-Labs/Nyx-API/components"
	nyxembed "github.com/Sumire-Labs/Nyx-API/embed"
	"github.com/Sumire-Labs/Nyx/database"
	"github.com/bwmarrin/discordgo"
)

func (b *Bot) handleTicketInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionMessageComponent {
		return
	}

	customID := i.MessageComponentData().CustomID

	switch customID {
	case "create_ticket":
		b.handleCreateTicket(s, i)
	case "close_ticket":
		b.handleCloseTicket(s, i)
	}
}

func (b *Bot) handleCreateTicket(s *discordgo.Session, i *discordgo.InteractionCreate) {
	userID := i.Member.User.ID
	guildID := i.GuildID

	openTickets, err := b.db.GetUserOpenTickets(guildID, userID)
	if err != nil {
		b.logger.Error("Failed to check user tickets: %v", err)
		b.respondError(s, i, "データベースエラーが発生しました")
		return
	}

	if len(openTickets) >= 3 {
		b.respondError(s, i, "同時に開けるチケットは最大3つまでです。既存のチケットを閉じてから再度お試しください。")
		return
	}

	guild, err := s.Guild(guildID)
	if err != nil {
		b.logger.Error("Failed to get guild: %v", err)
		b.respondError(s, i, "サーバー情報の取得に失敗しました")
		return
	}

	channelName := fmt.Sprintf("ticket-%s", userID)
	
	channel, err := s.GuildChannelCreate(guildID, &discordgo.GuildChannelCreateData{
		Name:     channelName,
		Type:     discordgo.ChannelTypeGuildText,
		Topic:    fmt.Sprintf("%s のサポートチケット", i.Member.User.Username),
		PermissionOverwrites: []*discordgo.PermissionOverwrite{
			{
				ID:   guildID,
				Type: discordgo.PermissionOverwriteTypeRole,
				Deny: discordgo.PermissionViewChannel,
			},
			{
				ID:    userID,
				Type:  discordgo.PermissionOverwriteTypeMember,
				Allow: discordgo.PermissionViewChannel | discordgo.PermissionSendMessages | discordgo.PermissionReadMessageHistory,
			},
		},
	})

	if err != nil {
		b.logger.Error("Failed to create ticket channel: %v", err)
		b.respondError(s, i, "チケットチャンネルの作成に失敗しました")
		return
	}

	for _, role := range guild.Roles {
		if role.Permissions&discordgo.PermissionManageChannels != 0 || role.Permissions&discordgo.PermissionAdministrator != 0 {
			_, err = s.ChannelPermissionSet(channel.ID, role.ID, discordgo.PermissionOverwriteTypeRole, 
				discordgo.PermissionViewChannel|discordgo.PermissionSendMessages|discordgo.PermissionReadMessageHistory, 0)
			if err != nil {
				b.logger.Warn("Failed to set admin role permissions for ticket: %v", err)
			}
		}
	}

	ticket := &database.Ticket{
		GuildID:   guildID,
		UserID:    userID,
		ChannelID: channel.ID,
		Title:     fmt.Sprintf("%s のサポートチケット", i.Member.User.Username),
		Status:    "open",
	}

	if err := b.db.CreateTicket(ticket); err != nil {
		b.logger.Error("Failed to save ticket to database: %v", err)
		s.ChannelDelete(channel.ID)
		b.respondError(s, i, "チケットの保存に失敗しました")
		return
	}

	embed := nyxembed.Success("✅ チケット作成完了", 
		fmt.Sprintf("チケットが作成されました: <#%s>\n\nサポートチームがまもなく対応いたします。", channel.ID)).
		SetFooter(nyxembed.Footer{
			Text: "チケットID: " + fmt.Sprintf("%d", ticket.ID),
		}).
		SetTimestamp(time.Now()).
		Build()

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
			Flags:  discordgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		b.logger.Error("Failed to respond to ticket creation: %v", err)
	}

	welcomeEmbed := nyxembed.Info("🎫 サポートチケット", 
		fmt.Sprintf("こんにちは <@%s> さん！\n\nサポートチケットにお越しいただきありがとうございます。\nお困りの内容について詳しく教えてください。\n\nサポートチームができるだけ早く対応いたします。", userID)).
		SetColor(0x5865F2).
		AddField(nyxembed.Field{
			Name:   "📝 チケット情報",
			Value:  fmt.Sprintf("**チケットID:** %d\n**作成日時:** %s", ticket.ID, time.Now().Format("2006-01-02 15:04:05")),
			Inline: false,
		}).
		SetTimestamp(time.Now()).
		Build()

	closeButton := nyxcomponents.DangerButton("🔒 チケットを閉じる", "close_ticket")
	actionRow := nyxcomponents.CreateButtonRow(closeButton)

	_, err = s.ChannelMessageSendComplex(channel.ID, &discordgo.MessageSend{
		Content:    fmt.Sprintf("<@%s>", userID),
		Embeds:     []*discordgo.MessageEmbed{welcomeEmbed},
		Components: []discordgo.MessageComponent{actionRow},
	})
	if err != nil {
		b.logger.Error("Failed to send welcome message: %v", err)
	}
}

func (b *Bot) handleCloseTicket(s *discordgo.Session, i *discordgo.InteractionCreate) {
	channelID := i.ChannelID
	userID := i.Member.User.ID

	ticket, err := b.db.GetTicket(channelID)
	if err != nil {
		b.logger.Error("Failed to get ticket: %v", err)
		b.respondError(s, i, "チケット情報の取得に失敗しました")
		return
	}

	if ticket == nil {
		b.respondError(s, i, "このチャンネルはチケットチャンネルではありません")
		return
	}

	if ticket.Status != "open" {
		b.respondError(s, i, "このチケットは既に閉じられています")
		return
	}

	member, err := s.GuildMember(i.GuildID, userID)
	if err != nil {
		b.logger.Error("Failed to get member: %v", err)
		b.respondError(s, i, "メンバー情報の取得に失敗しました")
		return
	}

	canClose := false
	if ticket.UserID == userID {
		canClose = true
	} else {
		for _, role := range member.Roles {
			roleObj, err := s.State.Role(i.GuildID, role)
			if err == nil && (roleObj.Permissions&discordgo.PermissionManageChannels != 0 || roleObj.Permissions&discordgo.PermissionAdministrator != 0) {
				canClose = true
				break
			}
		}
	}

	if !canClose {
		b.respondError(s, i, "チケットを閉じる権限がありません")
		return
	}

	if err := b.db.CloseTicket(channelID, userID); err != nil {
		b.logger.Error("Failed to close ticket in database: %v", err)
		b.respondError(s, i, "チケットのクローズに失敗しました")
		return
	}

	embed := nyxembed.Success("🔒 チケットクローズ", 
		fmt.Sprintf("チケットが <@%s> によってクローズされました。\n\n5秒後にチャンネルが削除されます。", userID)).
		SetFooter(nyxembed.Footer{
			Text: fmt.Sprintf("チケットID: %d", ticket.ID),
		}).
		SetTimestamp(time.Now()).
		Build()

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})
	if err != nil {
		b.logger.Error("Failed to respond to ticket closure: %v", err)
	}

	time.AfterFunc(5*time.Second, func() {
		_, err := s.ChannelDelete(channelID)
		if err != nil {
			b.logger.Error("Failed to delete ticket channel: %v", err)
		}
	})
}

func (b *Bot) respondError(s *discordgo.Session, i *discordgo.InteractionCreate, message string) {
	embed := nyxembed.Error("❌ エラー", message).Build()

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
			Flags:  discordgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		b.logger.Error("Failed to send error response: %v", err)
	}
}