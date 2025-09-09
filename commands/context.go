package commands

import (
	"fmt"
	
	nyxembed "github.com/Sumire-Labs/Nyx-API/embed"
	"github.com/Sumire-Labs/Nyx-API/logger"
	"github.com/Sumire-Labs/Nyx/database"
	"github.com/Sumire-Labs/Nyx/services"
	"github.com/bwmarrin/discordgo"
)

// DIContext DI対応の新しいContext実装
type DIContext struct {
	Session    *discordgo.Session
	Message    *discordgo.Message
	Args       []string
	Services   services.ServiceLocator
}

// DISlashContext DI対応の新しいSlashContext実装
type DISlashContext struct {
	Session     *discordgo.Session
	Interaction *discordgo.Interaction
	Services    services.ServiceLocator
}

// NewDIContext 新しいDIContextを作成
func NewDIContext(session *discordgo.Session, message *discordgo.Message, args []string, services services.ServiceLocator) *DIContext {
	return &DIContext{
		Session:  session,
		Message:  message,
		Args:     args,
		Services: services,
	}
}

// NewDISlashContext 新しいDISlashContextを作成
func NewDISlashContext(session *discordgo.Session, interaction *discordgo.Interaction, services services.ServiceLocator) *DISlashContext {
	return &DISlashContext{
		Session:     session,
		Interaction: interaction,
		Services:    services,
	}
}

// DIContext の services.Context インターフェース実装

func (c *DIContext) GetSession() *discordgo.Session {
	return c.Session
}

func (c *DIContext) GetMessage() *discordgo.Message {
	return c.Message
}

func (c *DIContext) GetArgs() []string {
	return c.Args
}

// 🔧 FIXED: エラーハンドリング追加（サービス取得失敗対応）
func (c *DIContext) GetBotConfig() services.BotConfigInfo {
	config, err := c.Services.BotConfig()
	if err != nil {
		// フォールバック: ログ出力後にnilを返す（呼び出し元で対応）
		fmt.Printf("ERROR: Failed to get BotConfig service: %v\n", err)
		return nil
	}
	return config
}

func (c *DIContext) GetLogger() services.LoggingService {
	logger, err := c.Services.Logger()
	if err != nil {
		// フォールバック: 標準出力でエラーを出力
		fmt.Printf("ERROR: Failed to get Logger service: %v\n", err)
		return nil
	}
	return logger
}

func (c *DIContext) GetDatabase() services.DatabaseService {
	db, err := c.Services.Database()
	if err != nil {
		// フォールバック: ログ出力後にnilを返す
		fmt.Printf("ERROR: Failed to get Database service: %v\n", err)
		return nil
	}
	return db
}

func (c *DIContext) Reply(content string) error {
	_, err := c.Session.ChannelMessageSend(c.Message.ChannelID, content)
	return err
}

func (c *DIContext) ReplyEmbed(embed *discordgo.MessageEmbed) error {
	_, err := c.Session.ChannelMessageSendEmbed(c.Message.ChannelID, embed)
	return err
}

func (c *DIContext) ReplyError(message string) error {
	embed := nyxembed.Error("エラー", message).Build()
	return c.ReplyEmbed(embed)
}

func (c *DIContext) ReplySuccess(message string) error {
	embed := nyxembed.Success("成功", message).Build()
	return c.ReplyEmbed(embed)
}

// DISlashContext の services.SlashContext インターフェース実装

func (sc *DISlashContext) GetSession() *discordgo.Session {
	return sc.Session
}

func (sc *DISlashContext) GetInteraction() *discordgo.Interaction {
	return sc.Interaction
}

// 🔧 FIXED: エラーハンドリング追加（SlashContext版）
func (sc *DISlashContext) GetBotConfig() services.BotConfigInfo {
	config, err := sc.Services.BotConfig()
	if err != nil {
		fmt.Printf("ERROR: Failed to get BotConfig service: %v\n", err)
		return nil
	}
	return config
}

func (sc *DISlashContext) GetLogger() services.LoggingService {
	logger, err := sc.Services.Logger()
	if err != nil {
		fmt.Printf("ERROR: Failed to get Logger service: %v\n", err)
		return nil
	}
	return logger
}

func (sc *DISlashContext) GetDatabase() services.DatabaseService {
	db, err := sc.Services.Database()
	if err != nil {
		fmt.Printf("ERROR: Failed to get Database service: %v\n", err)
		return nil
	}
	return db
}

func (sc *DISlashContext) Reply(content string, ephemeral bool) error {
	var flags discordgo.MessageFlags
	if ephemeral {
		flags = discordgo.MessageFlagsEphemeral
	}

	return sc.Session.InteractionRespond(sc.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
			Flags:   flags,
		},
	})
}

func (sc *DISlashContext) ReplyEmbed(embed *discordgo.MessageEmbed, ephemeral bool) error {
	var flags discordgo.MessageFlags
	if ephemeral {
		flags = discordgo.MessageFlagsEphemeral
	}

	return sc.Session.InteractionRespond(sc.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
			Flags:  flags,
		},
	})
}

func (sc *DISlashContext) ReplyError(message string, ephemeral bool) error {
	embed := nyxembed.Error("エラー", message).Build()
	return sc.ReplyEmbed(embed, ephemeral)
}

func (sc *DISlashContext) ReplySuccess(message string, ephemeral bool) error {
	embed := nyxembed.Success("成功", message).Build()
	return sc.ReplyEmbed(embed, ephemeral)
}

func (sc *DISlashContext) GetOptionValue(name string) *discordgo.ApplicationCommandInteractionDataOption {
	for _, option := range sc.Interaction.ApplicationCommandData().Options {
		if option.Name == name {
			return option
		}
	}
	return nil
}

// 後方互換性のためのアダプターメソッド群

// DB 旧Context互換のため、具体的なDatabase型を返す
func (c *DIContext) DB() *database.Database {
	// 🔧 FIXED: エラーハンドリング追加
	db, err := c.Services.Database()
	if err != nil {
		fmt.Printf("ERROR: Failed to get Database service in DB(): %v\n", err)
		return nil
	}
	if dbImpl, ok := db.(*database.Database); ok {
		return dbImpl
	}
	return nil
}

// Logger 旧Context互換のため、具体的なLogger型を返す
func (c *DIContext) Logger() *logger.Logger {
	// 🔧 FIXED: エラーハンドリング追加
	log, err := c.Services.Logger()
	if err != nil {
		fmt.Printf("ERROR: Failed to get Logger service in Logger(): %v\n", err)
		return nil
	}
	if logImpl, ok := log.(*logger.Logger); ok {
		return logImpl
	}
	return nil
}

// Bot 旧Context互換のため、BotInterface型を返す
func (c *DIContext) Bot() BotInterface {
	return &botConfigAdapter{config: c.GetBotConfig(), session: c.Session}
}

// 旧SlashContext用の後方互換メソッド

// DB 旧SlashContext互換のため、具体的なDatabase型を返す
func (sc *DISlashContext) DB() *database.Database {
	// 🔧 FIXED: エラーハンドリング追加
	db, err := sc.Services.Database()
	if err != nil {
		fmt.Printf("ERROR: Failed to get Database service in SlashContext DB(): %v\n", err)
		return nil
	}
	if dbImpl, ok := db.(*database.Database); ok {
		return dbImpl
	}
	return nil
}

// Logger 旧SlashContext互換のため、具体的なLogger型を返す
func (sc *DISlashContext) Logger() *logger.Logger {
	// 🔧 FIXED: エラーハンドリング追加
	log, err := sc.Services.Logger()
	if err != nil {
		fmt.Printf("ERROR: Failed to get Logger service in SlashContext Logger(): %v\n", err)
		return nil
	}
	if logImpl, ok := log.(*logger.Logger); ok {
		return logImpl
	}
	return nil
}

// Bot 旧SlashContext互換のため、BotInterface型を返す
func (sc *DISlashContext) Bot() BotInterface {
	return &botConfigAdapter{config: sc.GetBotConfig(), session: sc.Session}
}

// botConfigAdapter BotConfigInfoをBotInterfaceにアダプト
type botConfigAdapter struct {
	config  services.BotConfigInfo
	session *discordgo.Session
}

func (b *botConfigAdapter) GetPrefix() string {
	return b.config.GetPrefix()
}

func (b *botConfigAdapter) IsOwner(userID string) bool {
	return b.config.IsOwner(userID)
}

func (b *botConfigAdapter) GetSession() *discordgo.Session {
	return b.session
}

// DIContext を既存の Context に変換するアダプター
func (c *DIContext) AsLegacyContext() *Context {
	return &Context{
		Session: c.Session,
		Message: c.Message,
		Args:    c.Args,
		Bot:     &botConfigAdapter{config: c.GetBotConfig(), session: c.Session},
		Logger:  c.Logger(),
		DB:      c.DB(),
	}
}

// DISlashContext を既存の SlashContext に変換するアダプター
func (sc *DISlashContext) AsLegacySlashContext() *SlashContext {
	return &SlashContext{
		Session:     sc.Session,
		Interaction: sc.Interaction,
		Bot:         &botConfigAdapter{config: sc.GetBotConfig(), session: sc.Session},
		Logger:      sc.Logger(),
		DB:          sc.DB(),
	}
}