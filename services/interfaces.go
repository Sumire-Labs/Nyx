package services

import (
	"github.com/Sumire-Labs/Nyx/database"
	"github.com/bwmarrin/discordgo"
)

// DatabaseService データベース操作の抽象化インターフェース
type DatabaseService interface {
	// Guild operations
	GetGuild(guildID string) (*database.Guild, error)
	CreateGuild(guild *database.Guild) error
	UpdateGuild(guild *database.Guild) error
	DeleteGuild(guildID string) error

	// LogSettings operations
	GetLogSettings(guildID string) (*database.LogSettings, error)
	CreateLogSettings(settings *database.LogSettings) error
	UpdateLogSettings(settings *database.LogSettings) error
	DeleteLogSettings(guildID string) error

	// ServerLog operations
	LogServerEvent(log *database.ServerLog) error
	GetServerLogs(guildID string, limit int) ([]*database.ServerLog, error)

	// TicketPanel operations
	GetTicketPanel(messageID string) (*database.TicketPanel, error)
	CreateTicketPanel(panel *database.TicketPanel) error
	DeleteTicketPanel(messageID string) error
	GetAllTicketPanels() ([]*database.TicketPanel, error)

	// Ticket operations
	GetTicket(channelID string) (*database.Ticket, error)
	CreateTicket(ticket *database.Ticket) error
	UpdateTicket(ticket *database.Ticket) error
	DeleteTicket(channelID string) error
	GetUserTickets(userID string) ([]*database.Ticket, error)

	// Database lifecycle
	Close() error
}

// LoggingService ログ出力の抽象化インターフェース
type LoggingService interface {
	Debug(format string, args ...interface{})
	Info(format string, args ...interface{})
	Warn(format string, args ...interface{})
	Error(format string, args ...interface{})
	Fatal(format string, args ...interface{})
	SetFileOutput(filename string) error
	SetWebhook(webhookURL, username, avatarURL string)
}

// CommandService コマンド管理の抽象化インターフェース
type CommandService interface {
	Register(cmd interface{}) // commands.Command を使用
	GetCommand(name string) (interface{}, bool) // commands.Command を返す
	GetAll() map[string]interface{} // map[string]*commands.Command を返す
	Count() int
	GetByCategory(category string) []interface{} // []*commands.Command を返す
	RegisterDefaultCommands()
}

// Command は既存の commands.Command を使用するため、ここでは定義しない
// type Command = commands.Command として既存のものを使用

// BotService Bot操作の抽象化インターフェース
type BotService interface {
	GetPrefix() string
	IsOwner(userID string) bool
	GetSession() *discordgo.Session
	Start() error
	Stop() error
	IsReady() bool
}

// BotConfigInfo Botの設定情報のみ提供（循環依存回避用）
type BotConfigInfo interface {
	GetPrefix() string
	IsOwner(userID string) bool
}

// Context コマンド実行時のコンテキスト
type Context interface {
	GetSession() *discordgo.Session
	GetMessage() *discordgo.Message
	GetArgs() []string
	GetBotConfig() BotConfigInfo  // BotService から BotConfigInfo に変更
	GetLogger() LoggingService
	GetDatabase() DatabaseService
	Reply(content string) error
	ReplyEmbed(embed *discordgo.MessageEmbed) error
	ReplyError(message string) error
	ReplySuccess(message string) error
	
	// 後方互換性メソッド
	DB() interface{} // *database.Database を返す
	Logger() interface{} // *logger.Logger を返す
	Bot() interface{} // BotInterface を返す
}

// SlashContext スラッシュコマンド実行時のコンテキスト
type SlashContext interface {
	GetSession() *discordgo.Session
	GetInteraction() *discordgo.Interaction
	GetBotConfig() BotConfigInfo  // BotService から BotConfigInfo に変更
	GetLogger() LoggingService
	GetDatabase() DatabaseService
	Reply(content string, ephemeral bool) error
	ReplyEmbed(embed *discordgo.MessageEmbed, ephemeral bool) error
	ReplyError(message string, ephemeral bool) error
	ReplySuccess(message string, ephemeral bool) error
	GetOptionValue(name string) *discordgo.ApplicationCommandInteractionDataOption
	
	// 後方互換性メソッド
	DB() interface{} // *database.Database を返す
	Logger() interface{} // *logger.Logger を返す
	Bot() interface{} // BotInterface を返す
}

// ServiceLocator サービスへのアクセスを提供
// 🔧 FIXED: エラーを返すように変更（パニック除去）
type ServiceLocator interface {
	Database() (DatabaseService, error)
	Logger() (LoggingService, error)
	Commands() (CommandService, error)
	BotConfig() (BotConfigInfo, error)
}