package bot

import (
	"fmt"
	"strings"
	"time"

	"github.com/Sumire-Labs/Nyx-API/logger"
	"github.com/Sumire-Labs/Nyx/commands"
	"github.com/Sumire-Labs/Nyx/database"
	"github.com/Sumire-Labs/Nyx/services"
	"github.com/Sumire-Labs/Nyx/utils"
	"github.com/bwmarrin/discordgo"
)

type Bot struct {
	session      *discordgo.Session
	config       Config
	services     services.ServiceLocator  // DI サービスロケーター
	ready        bool
	
	// 🔧 FIXED: 無制限キャッシュ → LRUキャッシュ（メモリリーク修正）
	messageCache *utils.LRUCache  // メッセージID -> 内容
	memberCache  *utils.LRUCache  // guildID:userID -> Member
	cleanupStop  []chan<- bool    // クリーンアップ停止チャンネル
	
	// 後方互換性のためのキャッシュフィールド
	db           *database.Database
	commands     *commands.Registry
	logger       *logger.Logger
}

type Config struct {
	Token          string
	Prefix         string
	Status         string
	ActivityType   string
	ActivityName   string
	Database       *database.Database
	Commands       *commands.Registry
	Logger         *logger.Logger
	Services       services.ServiceLocator  // DI サービスロケーター
	SlashCommands  bool
	LoggingChannel string
	OwnerIDs       []string
}

func New(config Config) (*Bot, error) {
	session, err := discordgo.New("Bot " + config.Token)
	if err != nil {
		return nil, fmt.Errorf("failed to create Discord session: %w", err)
	}
	
	// 🔧 FIXED: LRUキャッシュでメモリ制限 (メッセージ: 1000件, 30分TTL)
	messageCache := utils.NewLRUCache(1000, 30*time.Minute)
	messageCacheStop := messageCache.StartCleanupRoutine(5 * time.Minute)
	
	// 🔧 FIXED: LRUキャッシュでメモリ制限 (メンバー: 500件, 1時間TTL)
	memberCache := utils.NewLRUCache(500, 1*time.Hour)
	memberCacheStop := memberCache.StartCleanupRoutine(10 * time.Minute)

	bot := &Bot{
		session:      session,
		config:       config,
		services:     config.Services,
		ready:        false,
		messageCache: messageCache,
		memberCache:  memberCache,
		cleanupStop:  []chan<- bool{messageCacheStop, memberCacheStop},
		// 後方互換性キャッシュ
		db:           config.Database,
		commands:     config.Commands,
		logger:       config.Logger,
	}
	
	session.AddHandler(bot.onReady)
	session.AddHandler(bot.onMessageCreate)
	session.AddHandler(bot.onInteractionCreate)
	
	// ログ用イベントハンドラー
	session.AddHandler(bot.handleGuildMemberAdd)
	session.AddHandler(bot.handleGuildMemberRemove)
	session.AddHandler(bot.handleGuildMemberUpdate)
	session.AddHandler(bot.handleMessageUpdate)
	session.AddHandler(bot.handleMessageDelete)
	session.AddHandler(bot.handleGuildBanAdd)
	session.AddHandler(bot.handleGuildBanRemove)
	session.AddHandler(bot.handleGuildRoleCreate)
	session.AddHandler(bot.handleGuildRoleUpdate)
	session.AddHandler(bot.handleGuildRoleDelete)
	session.AddHandler(bot.handleChannelCreate)
	session.AddHandler(bot.handleChannelUpdate)
	session.AddHandler(bot.handleChannelDelete)
	
	session.Identify.Intents = discordgo.IntentsAll
	
	return bot, nil
}

func (b *Bot) Start() error {
	if err := b.session.Open(); err != nil {
		return fmt.Errorf("failed to open Discord session: %w", err)
	}
	return nil
}

func (b *Bot) Stop() error {
	// 🔧 FIXED: キャッシュクリーンアップ停止（リソース適切解放）
	b.getLogger().Info("Stopping cache cleanup routines...")
	for _, stop := range b.cleanupStop {
		select {
		case stop <- true:
		default: // ノンブロッキング
		}
		close(stop)
	}
	
	if b.config.SlashCommands {
		b.getLogger().Info("Removing slash commands...")
		if err := b.removeSlashCommands(); err != nil {
			b.getLogger().Error("Failed to remove slash commands: %v", err)
		}
	}
	
	return b.session.Close()
}

func (b *Bot) onReady(s *discordgo.Session, r *discordgo.Ready) {
	b.getLogger().Info("Bot is ready! Logged in as %s#%s (%s)", r.User.Username, r.User.Discriminator, r.User.ID)
	
	if err := b.updateStatus(); err != nil {
		b.getLogger().Error("Failed to update status: %v", err)
	}
	
	if b.config.SlashCommands {
		if err := b.registerSlashCommands(); err != nil {
			b.getLogger().Error("Failed to register slash commands: %v", err)
		} else {
			b.getLogger().Info("Slash commands registered successfully")
		}
	}
	
	b.ready = true
}

func (b *Bot) onMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID || m.Author.Bot {
		return
	}
	
	// 🔧 FIXED: LRUキャッシュ使用（ログ用）
	if m.GuildID != "" {
		b.messageCache.Set(m.ID, m.Content)
		
		// メンバーキャッシュに保存
		if m.Member != nil {
			b.memberCache.Set(m.GuildID+":"+m.Author.ID, m.Member)
		}
	}
	
	if !strings.HasPrefix(m.Content, b.config.Prefix) {
		return
	}
	
	content := strings.TrimPrefix(m.Content, b.config.Prefix)
	parts := strings.Fields(content)
	if len(parts) == 0 {
		return
	}
	
	cmdName := strings.ToLower(parts[0])
	args := parts[1:]
	
	cmd, exists := b.getCommands().GetCommand(cmdName)
	if !exists {
		return
	}
	
	// DI対応の新しいContextを使用（後方互換性維持）
	var ctx interface{}
	if b.services != nil {
		// DI版Contextを使用
		ctx = commands.NewDIContext(s, m.Message, args, b.services)
	} else {
		// 既存版Contextを使用（後方互換性）
		ctx = &commands.Context{
			Session: s,
			Message: m.Message,
			Args:    args,
			Bot:     b.toBotInterface(),
			Logger:  b.getLogger(),
			DB:      b.getDatabase(),
		}
	}
	
	if cmd.OwnerOnly && !b.isOwner(m.Author.ID) {
		if diCtx, ok := ctx.(*commands.DIContext); ok {
			diCtx.ReplyError("このコマンドはBot所有者のみ実行できます。")
		} else if oldCtx, ok := ctx.(*commands.Context); ok {
			oldCtx.ReplyError("このコマンドはBot所有者のみ実行できます。")
		}
		return
	}
	
	if cmd.RequiredPermissions != 0 {
		perms, err := s.UserChannelPermissions(m.Author.ID, m.ChannelID)
		if err != nil {
			b.getLogger().Error("Failed to get user permissions: %v", err)
			return
		}
		
		if perms&cmd.RequiredPermissions != cmd.RequiredPermissions && perms&discordgo.PermissionAdministrator == 0 {
			if diCtx, ok := ctx.(*commands.DIContext); ok {
				diCtx.ReplyError("このコマンドを実行する権限がありません。")
			} else if oldCtx, ok := ctx.(*commands.Context); ok {
				oldCtx.ReplyError("このコマンドを実行する権限がありません。")
			}
			return
		}
	}
	
	b.getLogger().Info("Command executed: %s by %s in %s", cmdName, m.Author.ID, m.GuildID)
	
	var err error
	if diCtx, ok := ctx.(*commands.DIContext); ok {
		// DI Context を既存 Context に変換して実行
		err = cmd.Execute(diCtx.AsLegacyContext())
	} else if oldCtx, ok := ctx.(*commands.Context); ok {
		err = cmd.Execute(oldCtx)
	}
	
	if err != nil {
		b.getLogger().Error("Command error (%s): %v", cmdName, err)
		if diCtx, ok := ctx.(*commands.DIContext); ok {
			diCtx.ReplyError(fmt.Sprintf("コマンド実行中にエラーが発生しました: %v", err))
		} else if oldCtx, ok := ctx.(*commands.Context); ok {
			oldCtx.ReplyError(fmt.Sprintf("コマンド実行中にエラーが発生しました: %v", err))
		}
	}
}

func (b *Bot) onInteractionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		b.handleSlashCommand(s, i)
	case discordgo.InteractionMessageComponent:
		b.handleComponent(s, i)
	}
}

func (b *Bot) handleSlashCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	cmdName := i.ApplicationCommandData().Name
	cmd, exists := b.getCommands().GetCommand(cmdName)
	if !exists {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "コマンドが見つかりません。",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}
	
	// DI対応の新しいSlashContextを使用（後方互換性維持）
	var ctx interface{}
	if b.services != nil {
		// DI版SlashContextを使用
		ctx = commands.NewDISlashContext(s, i.Interaction, b.services)
	} else {
		// 既存版SlashContextを使用（後方互換性）
		ctx = &commands.SlashContext{
			Session:     s,
			Interaction: i.Interaction,
			Bot:         b.toBotInterface(),
			Logger:      b.getLogger(),
			DB:          b.getDatabase(),
		}
	}
	
	if cmd.OwnerOnly && !b.isOwner(i.Member.User.ID) {
		if diCtx, ok := ctx.(*commands.DISlashContext); ok {
			diCtx.ReplyError("このコマンドはBot所有者のみ実行できます。", true)
		} else if oldCtx, ok := ctx.(*commands.SlashContext); ok {
			oldCtx.ReplyError("このコマンドはBot所有者のみ実行できます。", true)
		}
		return
	}
	
	b.getLogger().Info("Slash command executed: %s by %s in %s", cmdName, i.Member.User.ID, i.GuildID)
	
	var err error
	if diCtx, ok := ctx.(*commands.DISlashContext); ok {
		// DI SlashContext を既存 SlashContext に変換して実行
		err = cmd.ExecuteSlash(diCtx.AsLegacySlashContext())
	} else if oldCtx, ok := ctx.(*commands.SlashContext); ok {
		err = cmd.ExecuteSlash(oldCtx)
	}
	
	if err != nil {
		b.getLogger().Error("Slash command error (%s): %v", cmdName, err)
		if diCtx, ok := ctx.(*commands.DISlashContext); ok {
			diCtx.ReplyError(fmt.Sprintf("コマンド実行中にエラーが発生しました: %v", err), true)
		} else if oldCtx, ok := ctx.(*commands.SlashContext); ok {
			oldCtx.ReplyError(fmt.Sprintf("コマンド実行中にエラーが発生しました: %v", err), true)
		}
	}
}

func (b *Bot) handleComponent(s *discordgo.Session, i *discordgo.InteractionCreate) {
	customID := i.MessageComponentData().CustomID
	
	if strings.HasPrefix(customID, "create_ticket") || strings.HasPrefix(customID, "close_ticket") {
		b.handleTicketInteraction(s, i)
		return
	}
	
	if strings.HasPrefix(customID, "page_") {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
		})
	}
}

func (b *Bot) updateStatus() error {
	var activityType discordgo.ActivityType
	switch b.config.ActivityType {
	case "playing":
		activityType = discordgo.ActivityTypeGame
	case "streaming":
		activityType = discordgo.ActivityTypeStreaming
	case "listening":
		activityType = discordgo.ActivityTypeListening
	case "watching":
		activityType = discordgo.ActivityTypeWatching
	case "competing":
		activityType = discordgo.ActivityTypeCompeting
	default:
		activityType = discordgo.ActivityTypeGame
	}
	
	var status discordgo.Status
	switch b.config.Status {
	case "online":
		status = discordgo.StatusOnline
	case "idle":
		status = discordgo.StatusIdle
	case "dnd":
		status = discordgo.StatusDoNotDisturb
	case "invisible":
		status = discordgo.StatusInvisible
	default:
		status = discordgo.StatusOnline
	}
	
	return b.session.UpdateStatusComplex(discordgo.UpdateStatusData{
		Activities: []*discordgo.Activity{
			{
				Name: b.config.ActivityName,
				Type: activityType,
			},
		},
		Status: string(status),
	})
}

func (b *Bot) registerSlashCommands() error {
	for _, cmd := range b.getCommands().GetAll() {
		if cmd.SlashCommand == nil {
			continue
		}
		
		_, err := b.session.ApplicationCommandCreate(b.session.State.User.ID, "", cmd.SlashCommand)
		if err != nil {
			return fmt.Errorf("failed to register slash command %s: %w", cmd.Name, err)
		}
	}
	return nil
}

func (b *Bot) removeSlashCommands() error {
	registeredCmds, err := b.session.ApplicationCommands(b.session.State.User.ID, "")
	if err != nil {
		return err
	}
	
	for _, cmd := range registeredCmds {
		if err := b.session.ApplicationCommandDelete(b.session.State.User.ID, "", cmd.ID); err != nil {
			b.getLogger().Error("Failed to delete slash command %s: %v", cmd.Name, err)
		}
	}
	return nil
}

func (b *Bot) isOwner(userID string) bool {
	for _, ownerID := range b.config.OwnerIDs {
		if userID == ownerID {
			return true
		}
	}
	return false
}

func (b *Bot) toBotInterface() commands.BotInterface {
	return &botInterface{bot: b}
}

type botInterface struct {
	bot *Bot
}

func (bi *botInterface) GetPrefix() string {
	return bi.bot.config.Prefix
}

func (bi *botInterface) IsOwner(userID string) bool {
	return bi.bot.isOwner(userID)
}

func (bi *botInterface) GetSession() *discordgo.Session {
	return bi.bot.session
}

// BotService インターフェースの実装メソッド

func (b *Bot) GetPrefix() string {
	return b.config.Prefix
}

func (b *Bot) IsOwner(userID string) bool {
	return b.isOwner(userID)
}

func (b *Bot) GetSession() *discordgo.Session {
	return b.session
}

func (b *Bot) IsReady() bool {
	return b.ready
}

// DIサービスアクセスメソッド

func (b *Bot) getDatabase() *database.Database {
	if b.services != nil {
		// 🔧 FIXED: エラーハンドリング追加
		db, err := b.services.Database()
		if err != nil {
			b.logger.Error("Failed to get database service: %v", err)
			return b.db // フォールバック
		}
		return db.(*database.Database)
	}
	return b.db // 後方互換性
}

func (b *Bot) getCommands() *commands.Registry {
	if b.services != nil {
		// 後方互換性のため、キャッシュされた Registry を使用
		if b.commands != nil {
			return b.commands
		}
		// フォールバック: 新しい Registry を作成
		registry := commands.NewRegistry()
		registry.RegisterDefaultCommands()
		return registry
	}
	return b.commands // 後方互換性
}

func (b *Bot) getLogger() *logger.Logger {
	if b.services != nil {
		// 🔧 FIXED: エラーハンドリング追加
		log, err := b.services.Logger()
		if err != nil {
			// フォールバックとして既存のloggerを使用
			if b.logger != nil {
				b.logger.Error("Failed to get logger service: %v", err)
				return b.logger
			}
			// 最終フォールバック: 新しいloggerを作成
			return logger.New("Nyx-Fallback", logger.InfoLevel)
		}
		return log.(*logger.Logger)
	}
	return b.logger // 後方互換性
}