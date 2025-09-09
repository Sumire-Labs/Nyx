package bot

import (
	"encoding/json"
	"fmt"
	"time"

	nyxembed "github.com/Sumire-Labs/Nyx-API/embed"
	"github.com/Sumire-Labs/Nyx/database"
	"github.com/bwmarrin/discordgo"
)

func (b *Bot) handleGuildMemberAdd(s *discordgo.Session, m *discordgo.GuildMemberAdd) {
	settings, err := b.getLogSettings(m.GuildID)
	if err != nil || settings == nil || !settings.Enabled || !settings.LogMemberJoin {
		return
	}

	embed := nyxembed.Success("👋 メンバー参加", "").
		SetColor(0x00ff00).
		AddField(nyxembed.Field{
			Name:   "👤 ユーザー",
			Value:  fmt.Sprintf("<@%s>\n`%s#%s` (ID: %s)", m.User.ID, m.User.Username, m.User.Discriminator, m.User.ID),
			Inline: false,
		}).
		AddField(nyxembed.Field{
			Name:   "📅 アカウント作成日",
			Value:  func() string {
				timestamp, _ := discordgo.SnowflakeTimestamp(m.User.ID)
				return timestamp.Format("2006/01/02 15:04:05")
			}(),
			Inline: true,
		}).
		AddField(nyxembed.Field{
			Name:   "⏰参加日時",
			Value:  time.Now().Format("2006/01/02 15:04:05"),
			Inline: true,
		}).
		SetThumbnail(nyxembed.Thumbnail{URL: m.User.AvatarURL("256")}).
		SetTimestamp(time.Now()).
		Build()

	b.sendLogMessage(settings, embed)

	metadata := fmt.Sprintf(`{"username": "%s", "discriminator": "%s"}`, m.User.Username, m.User.Discriminator)
	serverLog := &database.ServerLog{
		GuildID:   m.GuildID,
		EventType: "member_join",
		UserID:    &m.User.ID,
		Metadata:  &metadata,
	}
	b.db.LogServerEvent(serverLog)
}

func (b *Bot) handleGuildMemberRemove(s *discordgo.Session, m *discordgo.GuildMemberRemove) {
	settings, err := b.getLogSettings(m.GuildID)
	if err != nil || settings == nil || !settings.Enabled || !settings.LogMemberLeave {
		return
	}

	embed := nyxembed.Error("👋 メンバー退出", "").
		SetColor(0xff0000).
		AddField(nyxembed.Field{
			Name:   "👤 ユーザー",
			Value:  fmt.Sprintf("<@%s>\n`%s#%s` (ID: %s)", m.User.ID, m.User.Username, m.User.Discriminator, m.User.ID),
			Inline: false,
		}).
		AddField(nyxembed.Field{
			Name:   "⏰ 退出日時",
			Value:  time.Now().Format("2006/01/02 15:04:05"),
			Inline: true,
		}).
		SetThumbnail(nyxembed.Thumbnail{URL: m.User.AvatarURL("256")}).
		SetTimestamp(time.Now()).
		Build()

	b.sendLogMessage(settings, embed)

	metadata := fmt.Sprintf(`{"username": "%s", "discriminator": "%s"}`, m.User.Username, m.User.Discriminator)
	serverLog := &database.ServerLog{
		GuildID:   m.GuildID,
		EventType: "member_leave",
		UserID:    &m.User.ID,
		Metadata:  &metadata,
	}
	b.db.LogServerEvent(serverLog)
}

func (b *Bot) handleMessageUpdate(s *discordgo.Session, m *discordgo.MessageUpdate) {
	settings, err := b.getLogSettings(m.GuildID)
	if err != nil || settings == nil || !settings.Enabled || !settings.LogMessageEdit {
		return
	}

	if m.Author == nil || m.Author.Bot {
		return
	}

	// 🔧 FIXED: LRUキャッシュ使用
	oldMessageInterface, exists := b.messageCache.Get(m.ID)
	var oldMessage string
	if exists {
		oldMessage = oldMessageInterface.(string)
	}
	if !exists {
		return
	}

	embed := nyxembed.Warning("✏️ メッセージ編集", "").
		SetColor(0xffff00).
		AddField(nyxembed.Field{
			Name:   "👤 ユーザー",
			Value:  fmt.Sprintf("<@%s> (%s)", m.Author.ID, m.Author.Username),
			Inline: true,
		}).
		AddField(nyxembed.Field{
			Name:   "📍 チャンネル",
			Value:  fmt.Sprintf("<#%s>", m.ChannelID),
			Inline: true,
		}).
		AddField(nyxembed.Field{
			Name:   "🔗 メッセージリンク",
			Value:  fmt.Sprintf("[移動する](https://discord.com/channels/%s/%s/%s)", m.GuildID, m.ChannelID, m.ID),
			Inline: true,
		}).
		AddField(nyxembed.Field{
			Name:   "📝 編集前",
			Value:  truncateText(oldMessage, 1020),
			Inline: false,
		}).
		AddField(nyxembed.Field{
			Name:   "📝 編集後",
			Value:  truncateText(m.Content, 1020),
			Inline: false,
		}).
		SetTimestamp(time.Now()).
		Build()

	b.sendLogMessage(settings, embed)

	serverLog := &database.ServerLog{
		GuildID:    m.GuildID,
		EventType:  "message_edit",
		UserID:     &m.Author.ID,
		ChannelID:  &m.ChannelID,
		OldContent: &oldMessage,
		NewContent: &m.Content,
	}
	b.db.LogServerEvent(serverLog)

	// 🔧 FIXED: LRUキャッシュ使用
	b.messageCache.Set(m.ID, m.Content)
}

func (b *Bot) handleMessageDelete(s *discordgo.Session, m *discordgo.MessageDelete) {
	settings, err := b.getLogSettings(m.GuildID)
	if err != nil || settings == nil || !settings.Enabled || !settings.LogMessageDelete {
		return
	}

	// 🔧 FIXED: LRUキャッシュ使用
	contentInterface, exists := b.messageCache.Get(m.ID)
	var content string
	if exists {
		content = contentInterface.(string)
	}
	if !exists {
		content = "*メッセージの内容を取得できませんでした*"
	}

	embed := nyxembed.Error("🗑️ メッセージ削除", "").
		SetColor(0xff0000).
		AddField(nyxembed.Field{
			Name:   "📍 チャンネル",
			Value:  fmt.Sprintf("<#%s>", m.ChannelID),
			Inline: true,
		}).
		AddField(nyxembed.Field{
			Name:   "🆔 メッセージID",
			Value:  m.ID,
			Inline: true,
		}).
		AddField(nyxembed.Field{
			Name:   "⏰ 削除日時",
			Value:  time.Now().Format("2006/01/02 15:04:05"),
			Inline: true,
		}).
		AddField(nyxembed.Field{
			Name:   "📝 削除されたメッセージ",
			Value:  truncateText(content, 1020),
			Inline: false,
		}).
		SetTimestamp(time.Now()).
		Build()

	b.sendLogMessage(settings, embed)

	serverLog := &database.ServerLog{
		GuildID:    m.GuildID,
		EventType:  "message_delete",
		ChannelID:  &m.ChannelID,
		OldContent: &content,
	}
	b.db.LogServerEvent(serverLog)

	// 🔧 FIXED: LRUキャッシュ使用（削除はTTLで自動処理）
	b.messageCache.Delete(m.ID)
}

func (b *Bot) handleGuildMemberUpdate(s *discordgo.Session, m *discordgo.GuildMemberUpdate) {
	settings, err := b.getLogSettings(m.GuildID)
	if err != nil || settings == nil || !settings.Enabled || !settings.LogNicknameChange {
		return
	}

	// 🔧 FIXED: LRUキャッシュ使用
	oldMemberInterface, exists := b.memberCache.Get(m.GuildID + ":" + m.User.ID)
	var oldMember *discordgo.Member
	if exists {
		oldMember = oldMemberInterface.(*discordgo.Member)
	}
	if !exists {
		return
	}

	if oldMember.Nick == m.Nick {
		return
	}

	oldNick := "なし"
	newNick := "なし"
	if oldMember.Nick != "" {
		oldNick = oldMember.Nick
	}
	if m.Nick != "" {
		newNick = m.Nick
	}

	embed := nyxembed.Warning("✏️ ニックネーム変更", "").
		SetColor(0xffff00).
		AddField(nyxembed.Field{
			Name:   "👤 ユーザー",
			Value:  fmt.Sprintf("<@%s> (%s)", m.User.ID, m.User.Username),
			Inline: false,
		}).
		AddField(nyxembed.Field{
			Name:   "📝 変更前",
			Value:  oldNick,
			Inline: true,
		}).
		AddField(nyxembed.Field{
			Name:   "📝 変更後",
			Value:  newNick,
			Inline: true,
		}).
		SetThumbnail(nyxembed.Thumbnail{URL: m.User.AvatarURL("256")}).
		SetTimestamp(time.Now()).
		Build()

	b.sendLogMessage(settings, embed)

	serverLog := &database.ServerLog{
		GuildID:    m.GuildID,
		EventType:  "nickname_change",
		UserID:     &m.User.ID,
		OldContent: &oldNick,
		NewContent: &newNick,
	}
	b.db.LogServerEvent(serverLog)

	// 🔧 FIXED: LRUキャッシュ使用
	b.memberCache.Set(m.GuildID+":"+m.User.ID, &discordgo.Member{
		User: m.User,
		Nick: m.Nick,
		Roles: m.Roles,
		JoinedAt: m.JoinedAt,
	})
}

func (b *Bot) handleGuildBanAdd(s *discordgo.Session, m *discordgo.GuildBanAdd) {
	settings, err := b.getLogSettings(m.GuildID)
	if err != nil || settings == nil || !settings.Enabled || !settings.LogBan {
		return
	}

	embed := nyxembed.Error("🔨 メンバーBAN", "").
		SetColor(0xff0000).
		AddField(nyxembed.Field{
			Name:   "👤 ユーザー",
			Value:  fmt.Sprintf("<@%s>\n`%s#%s` (ID: %s)", m.User.ID, m.User.Username, m.User.Discriminator, m.User.ID),
			Inline: false,
		}).
		AddField(nyxembed.Field{
			Name:   "⏰ BAN日時",
			Value:  time.Now().Format("2006/01/02 15:04:05"),
			Inline: true,
		}).
		SetThumbnail(nyxembed.Thumbnail{URL: m.User.AvatarURL("256")}).
		SetTimestamp(time.Now()).
		Build()

	b.sendLogMessage(settings, embed)

	metadata := fmt.Sprintf(`{"username": "%s", "discriminator": "%s"}`, m.User.Username, m.User.Discriminator)
	serverLog := &database.ServerLog{
		GuildID:   m.GuildID,
		EventType: "ban",
		UserID:    &m.User.ID,
		Metadata:  &metadata,
	}
	b.db.LogServerEvent(serverLog)
}

func (b *Bot) handleGuildBanRemove(s *discordgo.Session, m *discordgo.GuildBanRemove) {
	settings, err := b.getLogSettings(m.GuildID)
	if err != nil || settings == nil || !settings.Enabled || !settings.LogUnban {
		return
	}

	embed := nyxembed.Success("🔓 BAN解除", "").
		SetColor(0x00ff00).
		AddField(nyxembed.Field{
			Name:   "👤 ユーザー",
			Value:  fmt.Sprintf("<@%s>\n`%s#%s` (ID: %s)", m.User.ID, m.User.Username, m.User.Discriminator, m.User.ID),
			Inline: false,
		}).
		AddField(nyxembed.Field{
			Name:   "⏰ BAN解除日時",
			Value:  time.Now().Format("2006/01/02 15:04:05"),
			Inline: true,
		}).
		SetThumbnail(nyxembed.Thumbnail{URL: m.User.AvatarURL("256")}).
		SetTimestamp(time.Now()).
		Build()

	b.sendLogMessage(settings, embed)

	metadata := fmt.Sprintf(`{"username": "%s", "discriminator": "%s"}`, m.User.Username, m.User.Discriminator)
	serverLog := &database.ServerLog{
		GuildID:   m.GuildID,
		EventType: "unban",
		UserID:    &m.User.ID,
		Metadata:  &metadata,
	}
	b.db.LogServerEvent(serverLog)
}

func (b *Bot) handleGuildRoleCreate(s *discordgo.Session, m *discordgo.GuildRoleCreate) {
	settings, err := b.getLogSettings(m.GuildID)
	if err != nil || settings == nil || !settings.Enabled || !settings.LogRoleCreate {
		return
	}

	embed := nyxembed.Success("🏷️ ロール作成", "").
		SetColor(0x00ff00).
		AddField(nyxembed.Field{
			Name:   "🏷️ ロール",
			Value:  fmt.Sprintf("<@&%s>\n`%s` (ID: %s)", m.Role.ID, m.Role.Name, m.Role.ID),
			Inline: false,
		}).
		AddField(nyxembed.Field{
			Name:   "🎨 色",
			Value:  fmt.Sprintf("#%06x", m.Role.Color),
			Inline: true,
		}).
		AddField(nyxembed.Field{
			Name:   "📍 位置",
			Value:  fmt.Sprintf("%d", m.Role.Position),
			Inline: true,
		}).
		AddField(nyxembed.Field{
			Name:   "⏰ 作成日時",
			Value:  time.Now().Format("2006/01/02 15:04:05"),
			Inline: true,
		}).
		SetTimestamp(time.Now()).
		Build()

	b.sendLogMessage(settings, embed)

	roleData, _ := json.Marshal(map[string]interface{}{
		"name":     m.Role.Name,
		"color":    m.Role.Color,
		"position": m.Role.Position,
	})
	metadata := string(roleData)

	serverLog := &database.ServerLog{
		GuildID:   m.GuildID,
		EventType: "role_create",
		RoleID:    &m.Role.ID,
		Metadata:  &metadata,
	}
	b.db.LogServerEvent(serverLog)
}

func (b *Bot) handleGuildRoleUpdate(s *discordgo.Session, m *discordgo.GuildRoleUpdate) {
	settings, err := b.getLogSettings(m.GuildID)
	if err != nil || settings == nil || !settings.Enabled || !settings.LogRoleUpdate {
		return
	}

	embed := nyxembed.Warning("✏️ ロール更新", "").
		SetColor(0xffff00).
		AddField(nyxembed.Field{
			Name:   "🏷️ ロール",
			Value:  fmt.Sprintf("<@&%s>\n`%s` (ID: %s)", m.Role.ID, m.Role.Name, m.Role.ID),
			Inline: false,
		}).
		AddField(nyxembed.Field{
			Name:   "⏰ 更新日時",
			Value:  time.Now().Format("2006/01/02 15:04:05"),
			Inline: true,
		}).
		SetTimestamp(time.Now()).
		Build()

	b.sendLogMessage(settings, embed)

	roleData, _ := json.Marshal(map[string]interface{}{
		"name":     m.Role.Name,
		"color":    m.Role.Color,
		"position": m.Role.Position,
	})
	metadata := string(roleData)

	serverLog := &database.ServerLog{
		GuildID:   m.GuildID,
		EventType: "role_update",
		RoleID:    &m.Role.ID,
		Metadata:  &metadata,
	}
	b.db.LogServerEvent(serverLog)
}

func (b *Bot) handleGuildRoleDelete(s *discordgo.Session, m *discordgo.GuildRoleDelete) {
	settings, err := b.getLogSettings(m.GuildID)
	if err != nil || settings == nil || !settings.Enabled || !settings.LogRoleDelete {
		return
	}

	embed := nyxembed.Error("🗑️ ロール削除", "").
		SetColor(0xff0000).
		AddField(nyxembed.Field{
			Name:   "🆔 ロールID",
			Value:  m.RoleID,
			Inline: true,
		}).
		AddField(nyxembed.Field{
			Name:   "⏰ 削除日時",
			Value:  time.Now().Format("2006/01/02 15:04:05"),
			Inline: true,
		}).
		SetTimestamp(time.Now()).
		Build()

	b.sendLogMessage(settings, embed)

	serverLog := &database.ServerLog{
		GuildID:   m.GuildID,
		EventType: "role_delete",
		RoleID:    &m.RoleID,
	}
	b.db.LogServerEvent(serverLog)
}

func (b *Bot) handleChannelCreate(s *discordgo.Session, m *discordgo.ChannelCreate) {
	settings, err := b.getLogSettings(m.GuildID)
	if err != nil || settings == nil || !settings.Enabled || !settings.LogChannelCreate {
		return
	}

	channelType := "不明"
	switch m.Type {
	case discordgo.ChannelTypeGuildText:
		channelType = "テキストチャンネル"
	case discordgo.ChannelTypeGuildVoice:
		channelType = "ボイスチャンネル"
	case discordgo.ChannelTypeGuildCategory:
		channelType = "カテゴリ"
	case discordgo.ChannelTypeGuildNews:
		channelType = "アナウンスチャンネル"
	case discordgo.ChannelTypeGuildForum:
		channelType = "フォーラムチャンネル"
	}

	embed := nyxembed.Success("📁 チャンネル作成", "").
		SetColor(0x00ff00).
		AddField(nyxembed.Field{
			Name:   "📁 チャンネル",
			Value:  fmt.Sprintf("<#%s>\n`%s` (ID: %s)", m.ID, m.Name, m.ID),
			Inline: false,
		}).
		AddField(nyxembed.Field{
			Name:   "📋 タイプ",
			Value:  channelType,
			Inline: true,
		}).
		AddField(nyxembed.Field{
			Name:   "⏰ 作成日時",
			Value:  time.Now().Format("2006/01/02 15:04:05"),
			Inline: true,
		}).
		SetTimestamp(time.Now()).
		Build()

	b.sendLogMessage(settings, embed)

	channelData, _ := json.Marshal(map[string]interface{}{
		"name": m.Name,
		"type": channelType,
	})
	metadata := string(channelData)

	serverLog := &database.ServerLog{
		GuildID:   m.GuildID,
		EventType: "channel_create",
		ChannelID: &m.ID,
		Metadata:  &metadata,
	}
	b.db.LogServerEvent(serverLog)
}

func (b *Bot) handleChannelUpdate(s *discordgo.Session, m *discordgo.ChannelUpdate) {
	settings, err := b.getLogSettings(m.GuildID)
	if err != nil || settings == nil || !settings.Enabled || !settings.LogChannelUpdate {
		return
	}

	embed := nyxembed.Warning("✏️ チャンネル更新", "").
		SetColor(0xffff00).
		AddField(nyxembed.Field{
			Name:   "📁 チャンネル",
			Value:  fmt.Sprintf("<#%s>\n`%s` (ID: %s)", m.ID, m.Name, m.ID),
			Inline: false,
		}).
		AddField(nyxembed.Field{
			Name:   "⏰ 更新日時",
			Value:  time.Now().Format("2006/01/02 15:04:05"),
			Inline: true,
		}).
		SetTimestamp(time.Now()).
		Build()

	b.sendLogMessage(settings, embed)

	channelData, _ := json.Marshal(map[string]interface{}{
		"name": m.Name,
	})
	metadata := string(channelData)

	serverLog := &database.ServerLog{
		GuildID:   m.GuildID,
		EventType: "channel_update",
		ChannelID: &m.ID,
		Metadata:  &metadata,
	}
	b.db.LogServerEvent(serverLog)
}

func (b *Bot) handleChannelDelete(s *discordgo.Session, m *discordgo.ChannelDelete) {
	settings, err := b.getLogSettings(m.GuildID)
	if err != nil || settings == nil || !settings.Enabled || !settings.LogChannelDelete {
		return
	}

	embed := nyxembed.Error("🗑️ チャンネル削除", "").
		SetColor(0xff0000).
		AddField(nyxembed.Field{
			Name:   "📁 チャンネル",
			Value:  fmt.Sprintf("`%s` (ID: %s)", m.Name, m.ID),
			Inline: false,
		}).
		AddField(nyxembed.Field{
			Name:   "⏰ 削除日時",
			Value:  time.Now().Format("2006/01/02 15:04:05"),
			Inline: true,
		}).
		SetTimestamp(time.Now()).
		Build()

	b.sendLogMessage(settings, embed)

	channelData, _ := json.Marshal(map[string]interface{}{
		"name": m.Name,
	})
	metadata := string(channelData)

	serverLog := &database.ServerLog{
		GuildID:   m.GuildID,
		EventType: "channel_delete",
		ChannelID: &m.ID,
		Metadata:  &metadata,
	}
	b.db.LogServerEvent(serverLog)
}

// ヘルパー関数
func (b *Bot) getLogSettings(guildID string) (*database.LogSettings, error) {
	return b.db.GetLogSettings(guildID)
}

func (b *Bot) sendLogMessage(settings *database.LogSettings, embed *discordgo.MessageEmbed) {
	if settings.LogChannelID == nil {
		return
	}

	b.session.ChannelMessageSendEmbed(*settings.LogChannelID, embed)
}

func truncateText(text string, maxLength int) string {
	if len(text) <= maxLength {
		return text
	}
	return text[:maxLength-3] + "..."
}