package bot

import (
	"fmt"
	"strings"

	"github.com/Sumire-Labs/Nyx-API/logger"
	"github.com/Sumire-Labs/Nyx/commands"
	"github.com/Sumire-Labs/Nyx/database"
	"github.com/bwmarrin/discordgo"
)

type Bot struct {
	session        *discordgo.Session
	config         Config
	db             *database.Database
	commands       *commands.Registry
	logger         *logger.Logger
	ready          bool
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
	SlashCommands  bool
	LoggingChannel string
	OwnerIDs       []string
}

func New(config Config) (*Bot, error) {
	session, err := discordgo.New("Bot " + config.Token)
	if err != nil {
		return nil, fmt.Errorf("failed to create Discord session: %w", err)
	}
	
	bot := &Bot{
		session:  session,
		config:   config,
		db:       config.Database,
		commands: config.Commands,
		logger:   config.Logger,
		ready:    false,
	}
	
	session.AddHandler(bot.onReady)
	session.AddHandler(bot.onMessageCreate)
	session.AddHandler(bot.onInteractionCreate)
	
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
	if b.config.SlashCommands {
		b.logger.Info("Removing slash commands...")
		if err := b.removeSlashCommands(); err != nil {
			b.logger.Error("Failed to remove slash commands: %v", err)
		}
	}
	
	return b.session.Close()
}

func (b *Bot) onReady(s *discordgo.Session, r *discordgo.Ready) {
	b.logger.Info("Bot is ready! Logged in as %s#%s (%s)", r.User.Username, r.User.Discriminator, r.User.ID)
	
	if err := b.updateStatus(); err != nil {
		b.logger.Error("Failed to update status: %v", err)
	}
	
	if b.config.SlashCommands {
		if err := b.registerSlashCommands(); err != nil {
			b.logger.Error("Failed to register slash commands: %v", err)
		} else {
			b.logger.Info("Slash commands registered successfully")
		}
	}
	
	b.ready = true
}

func (b *Bot) onMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID || m.Author.Bot {
		return
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
	
	cmd, exists := b.commands.GetCommand(cmdName)
	if !exists {
		return
	}
	
	ctx := &commands.Context{
		Session: s,
		Message: m.Message,
		Args:    args,
		Bot:     b.toBotInterface(),
		Logger:  b.logger,
		DB:      b.db,
	}
	
	if cmd.OwnerOnly && !b.isOwner(m.Author.ID) {
		ctx.ReplyError("このコマンドはBot所有者のみ実行できます。")
		return
	}
	
	if cmd.RequiredPermissions != 0 {
		perms, err := s.UserChannelPermissions(m.Author.ID, m.ChannelID)
		if err != nil {
			b.logger.Error("Failed to get user permissions: %v", err)
			return
		}
		
		if perms&cmd.RequiredPermissions != cmd.RequiredPermissions && perms&discordgo.PermissionAdministrator == 0 {
			ctx.ReplyError("このコマンドを実行する権限がありません。")
			return
		}
	}
	
	b.logger.Info("Command executed: %s by %s in %s", cmdName, m.Author.ID, m.GuildID)
	
	if err := cmd.Execute(ctx); err != nil {
		b.logger.Error("Command error (%s): %v", cmdName, err)
		ctx.ReplyError(fmt.Sprintf("コマンド実行中にエラーが発生しました: %v", err))
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
	cmd, exists := b.commands.GetCommand(cmdName)
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
	
	ctx := &commands.SlashContext{
		Session:     s,
		Interaction: i.Interaction,
		Bot:         b.toBotInterface(),
		Logger:      b.logger,
		DB:          b.db,
	}
	
	if cmd.OwnerOnly && !b.isOwner(i.Member.User.ID) {
		ctx.ReplyError("このコマンドはBot所有者のみ実行できます。", true)
		return
	}
	
	b.logger.Info("Slash command executed: %s by %s in %s", cmdName, i.Member.User.ID, i.GuildID)
	
	if err := cmd.ExecuteSlash(ctx); err != nil {
		b.logger.Error("Slash command error (%s): %v", cmdName, err)
		ctx.ReplyError(fmt.Sprintf("コマンド実行中にエラーが発生しました: %v", err), true)
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
	for _, cmd := range b.commands.GetAll() {
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
			b.logger.Error("Failed to delete slash command %s: %v", cmd.Name, err)
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