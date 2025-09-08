package commands

import (
	nyxembed "github.com/Sumire-Labs/Nyx-API/embed"
	"github.com/Sumire-Labs/Nyx-API/logger"
	"github.com/Sumire-Labs/Nyx/database"
	"github.com/bwmarrin/discordgo"
)

type Registry struct {
	commands map[string]*Command
}

type Command struct {
	Name                string
	Description         string
	Usage               string
	Aliases             []string
	Category            string
	RequiredPermissions int64
	OwnerOnly           bool
	GuildOnly           bool
	SlashCommand        *discordgo.ApplicationCommand
	Execute             func(*Context) error
	ExecuteSlash        func(*SlashContext) error
}

type Context struct {
	Session *discordgo.Session
	Message *discordgo.Message
	Args    []string
	Bot     BotInterface
	Logger  *logger.Logger
	DB      *database.Database
}

type SlashContext struct {
	Session     *discordgo.Session
	Interaction *discordgo.Interaction
	Bot         BotInterface
	Logger      *logger.Logger
	DB          *database.Database
}

type BotInterface interface {
	GetPrefix() string
	IsOwner(userID string) bool
	GetSession() *discordgo.Session
}

func NewRegistry() *Registry {
	return &Registry{
		commands: make(map[string]*Command),
	}
}

func (r *Registry) Register(cmd *Command) {
	r.commands[cmd.Name] = cmd

	for _, alias := range cmd.Aliases {
		r.commands[alias] = cmd
	}
}

func (r *Registry) GetCommand(name string) (*Command, bool) {
	cmd, exists := r.commands[name]
	return cmd, exists
}

func (r *Registry) GetAll() map[string]*Command {
	unique := make(map[string]*Command)
	for _, cmd := range r.commands {
		unique[cmd.Name] = cmd
	}
	return unique
}

func (r *Registry) Count() int {
	return len(r.GetAll())
}

func (r *Registry) GetByCategory(category string) []*Command {
	var commands []*Command
	seen := make(map[string]bool)

	for _, cmd := range r.commands {
		if cmd.Category == category && !seen[cmd.Name] {
			commands = append(commands, cmd)
			seen[cmd.Name] = true
		}
	}
	return commands
}

var DefaultCommands []*Command

func (r *Registry) RegisterDefaultCommands() {
	for _, cmd := range DefaultCommands {
		r.Register(cmd)
	}
}

func (c *Context) Reply(content string) error {
	_, err := c.Session.ChannelMessageSend(c.Message.ChannelID, content)
	return err
}

func (c *Context) ReplyEmbed(embed *discordgo.MessageEmbed) error {
	_, err := c.Session.ChannelMessageSendEmbed(c.Message.ChannelID, embed)
	return err
}

func (c *Context) ReplyError(message string) error {
	embed := nyxembed.Error("エラー", message).Build()
	return c.ReplyEmbed(embed)
}

func (c *Context) ReplySuccess(message string) error {
	embed := nyxembed.Success("成功", message).Build()
	return c.ReplyEmbed(embed)
}

func (sc *SlashContext) Reply(content string, ephemeral bool) error {
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

func (sc *SlashContext) ReplyEmbed(embed *discordgo.MessageEmbed, ephemeral bool) error {
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

func (sc *SlashContext) ReplyError(message string, ephemeral bool) error {
	embed := nyxembed.Error("エラー", message).Build()
	return sc.ReplyEmbed(embed, ephemeral)
}

func (sc *SlashContext) ReplySuccess(message string, ephemeral bool) error {
	embed := nyxembed.Success("成功", message).Build()
	return sc.ReplyEmbed(embed, ephemeral)
}

func (sc *SlashContext) GetOptionValue(name string) *discordgo.ApplicationCommandInteractionDataOption {
	for _, option := range sc.Interaction.ApplicationCommandData().Options {
		if option.Name == name {
			return option
		}
	}
	return nil
}

// インターフェース適合用のアダプターメソッド群

// Context のアダプターメソッド
func (c *Context) GetSession() *discordgo.Session {
	return c.Session
}

func (c *Context) GetMessage() *discordgo.Message {
	return c.Message
}

func (c *Context) GetArgs() []string {
	return c.Args
}

func (c *Context) GetBot() interface{} { // services.BotService
	return c.Bot
}

func (c *Context) GetLogger() interface{} { // services.LoggingService
	return c.Logger
}

func (c *Context) GetDatabase() interface{} { // services.DatabaseService
	return c.DB
}

// 後方互換性メソッド - 既存コードで使用されるメソッド名
func (c *Context) GetDB() *database.Database {
	return c.DB
}

func (c *Context) GetLogger2() *logger.Logger {
	return c.Logger
}

func (c *Context) GetBotInterface() BotInterface {
	return c.Bot
}

// SlashContext のアダプターメソッド
func (sc *SlashContext) GetSession() *discordgo.Session {
	return sc.Session
}

func (sc *SlashContext) GetInteraction() *discordgo.Interaction {
	return sc.Interaction
}

func (sc *SlashContext) GetBot() interface{} { // services.BotService
	return sc.Bot
}

func (sc *SlashContext) GetLogger() interface{} { // services.LoggingService
	return sc.Logger
}

func (sc *SlashContext) GetDatabase() interface{} { // services.DatabaseService
	return sc.DB
}

// 後方互換性メソッド - 既存コードで使用されるメソッド名
func (sc *SlashContext) GetDB() *database.Database {
	return sc.DB
}

func (sc *SlashContext) GetLogger2() *logger.Logger {
	return sc.Logger
}

func (sc *SlashContext) GetBotInterface() BotInterface {
	return sc.Bot
}

// CommandService インターフェースの実装 (Registry 用)

// Register のインターフェース適合用ラッパー
func (r *Registry) RegisterInterface(cmd interface{}) {
	if command, ok := cmd.(*Command); ok {
		r.Register(command)
	}
}

// GetCommand のインターフェース適合用ラッパー
func (r *Registry) GetCommandInterface(name string) (interface{}, bool) {
	cmd, exists := r.GetCommand(name)
	return cmd, exists
}

// GetAll のインターフェース適合用ラッパー
func (r *Registry) GetAllInterface() map[string]interface{} {
	all := r.GetAll()
	result := make(map[string]interface{})
	for k, v := range all {
		result[k] = v
	}
	return result
}

// GetByCategory のインターフェース適合用ラッパー
func (r *Registry) GetByCategoryInterface(category string) []interface{} {
	cmds := r.GetByCategory(category)
	result := make([]interface{}, len(cmds))
	for i, cmd := range cmds {
		result[i] = cmd
	}
	return result
}
