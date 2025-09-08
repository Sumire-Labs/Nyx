package commands

import (
	"fmt"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/Sumire-Labs/Nyx-API/embed"
	"github.com/Sumire-Labs/Nyx-API/format"
	"github.com/bwmarrin/discordgo"
)

func (r *Registry) pingCommand(ctx *Context) error {
	start := time.Now()
	
	msg, err := ctx.Session.ChannelMessageSend(ctx.Message.ChannelID, "計測中...")
	if err != nil {
		return err
	}
	
	latency := time.Since(start)
	heartbeat := ctx.Session.HeartbeatLatency()
	
	embed := embed.Info("🏓 Pong!", fmt.Sprintf(
		"**Bot レイテンシ:** %dms\n**WebSocket レイテンシ:** %dms",
		latency.Milliseconds(),
		heartbeat.Milliseconds(),
	)).Build()
	
	_, err = ctx.Session.ChannelMessageEditEmbed(ctx.Message.ChannelID, msg.ID, embed)
	return err
}

func (r *Registry) pingSlashCommand(ctx *SlashContext) error {
	start := time.Now()
	
	err := ctx.Reply("計測中...", false)
	if err != nil {
		return err
	}
	
	latency := time.Since(start)
	heartbeat := ctx.Session.HeartbeatLatency()
	
	embed := embed.Info("🏓 Pong!", fmt.Sprintf(
		"**Bot レイテンシ:** %dms\n**WebSocket レイテンシ:** %dms",
		latency.Milliseconds(),
		heartbeat.Milliseconds(),
	)).Build()
	
	_, err = ctx.Session.InteractionResponseEdit(ctx.Interaction, &discordgo.WebhookEdit{
		Embeds: &[]*discordgo.MessageEmbed{embed},
	})
	return err
}

func (r *Registry) helpCommand(ctx *Context) error {
	if len(ctx.Args) > 0 {
		return r.helpSpecificCommand(ctx, ctx.Args[0])
	}
	
	return r.helpAllCommands(ctx)
}

func (r *Registry) helpSlashCommand(ctx *SlashContext) error {
	option := ctx.GetOptionValue("command")
	if option != nil {
		return r.helpSpecificCommandSlash(ctx, option.StringValue())
	}
	
	return r.helpAllCommandsSlash(ctx)
}

func (r *Registry) helpAllCommands(ctx *Context) error {
	categories := make(map[string][]*Command)
	
	for _, cmd := range r.GetAll() {
		if cmd.Category == "" {
			cmd.Category = "その他"
		}
		categories[cmd.Category] = append(categories[cmd.Category], cmd)
	}
	
	embed := embed.New().
		SetTitle("📚 コマンド一覧").
		SetDescription(fmt.Sprintf("プレフィックス: `%s`\n詳細は `%shelp [コマンド名]` で確認できます。", 
			ctx.Bot.GetPrefix(), ctx.Bot.GetPrefix())).
		SetColor(embed.ColorInfo)
	
	var categoryNames []string
	for category := range categories {
		categoryNames = append(categoryNames, category)
	}
	sort.Strings(categoryNames)
	
	for _, category := range categoryNames {
		commands := categories[category]
		var cmdNames []string
		for _, cmd := range commands {
			cmdNames = append(cmdNames, fmt.Sprintf("`%s`", cmd.Name))
		}
		
		embed.AddField(embed.Field{
			Name:   fmt.Sprintf("📁 %s (%d)", category, len(commands)),
			Value:  strings.Join(cmdNames, ", "),
			Inline: false,
		})
	}
	
	embed.SetFooter(embed.Footer{
		Text: fmt.Sprintf("合計 %d コマンド", r.Count()),
	})
	
	return ctx.ReplyEmbed(embed.Build())
}

func (r *Registry) helpAllCommandsSlash(ctx *SlashContext) error {
	categories := make(map[string][]*Command)
	
	for _, cmd := range r.GetAll() {
		if cmd.Category == "" {
			cmd.Category = "その他"
		}
		categories[cmd.Category] = append(categories[cmd.Category], cmd)
	}
	
	embed := embed.New().
		SetTitle("📚 コマンド一覧").
		SetDescription(fmt.Sprintf("プレフィックス: `%s`\n詳細は `/help [コマンド名]` で確認できます。", 
			ctx.Bot.GetPrefix())).
		SetColor(embed.ColorInfo)
	
	var categoryNames []string
	for category := range categories {
		categoryNames = append(categoryNames, category)
	}
	sort.Strings(categoryNames)
	
	for _, category := range categoryNames {
		commands := categories[category]
		var cmdNames []string
		for _, cmd := range commands {
			cmdNames = append(cmdNames, fmt.Sprintf("`%s`", cmd.Name))
		}
		
		embed.AddField(embed.Field{
			Name:   fmt.Sprintf("📁 %s (%d)", category, len(commands)),
			Value:  strings.Join(cmdNames, ", "),
			Inline: false,
		})
	}
	
	embed.SetFooter(embed.Footer{
		Text: fmt.Sprintf("合計 %d コマンド", r.Count()),
	})
	
	return ctx.ReplyEmbed(embed.Build(), false)
}

func (r *Registry) helpSpecificCommand(ctx *Context, cmdName string) error {
	cmd, exists := r.GetCommand(cmdName)
	if !exists {
		return ctx.ReplyError(fmt.Sprintf("コマンド `%s` は存在しません。", cmdName))
	}
	
	embed := embed.New().
		SetTitle(fmt.Sprintf("📖 %s コマンド", cmd.Name)).
		SetDescription(cmd.Description).
		SetColor(embed.ColorInfo)
	
	embed.AddField(embed.Field{
		Name:   "使用方法",
		Value:  fmt.Sprintf("`%s%s`", ctx.Bot.GetPrefix(), cmd.Usage),
		Inline: false,
	})
	
	if len(cmd.Aliases) > 0 {
		aliases := make([]string, len(cmd.Aliases))
		for i, alias := range cmd.Aliases {
			aliases[i] = fmt.Sprintf("`%s`", alias)
		}
		embed.AddField(embed.Field{
			Name:   "エイリアス",
			Value:  strings.Join(aliases, ", "),
			Inline: true,
		})
	}
	
	embed.AddField(embed.Field{
		Name:   "カテゴリ",
		Value:  cmd.Category,
		Inline: true,
	})
	
	var flags []string
	if cmd.GuildOnly {
		flags = append(flags, "サーバー限定")
	}
	if cmd.OwnerOnly {
		flags = append(flags, "所有者限定")
	}
	if cmd.RequiredPermissions != 0 {
		flags = append(flags, "権限必要")
	}
	
	if len(flags) > 0 {
		embed.AddField(embed.Field{
			Name:   "制限",
			Value:  strings.Join(flags, ", "),
			Inline: true,
		})
	}
	
	return ctx.ReplyEmbed(embed.Build())
}

func (r *Registry) helpSpecificCommandSlash(ctx *SlashContext, cmdName string) error {
	cmd, exists := r.GetCommand(cmdName)
	if !exists {
		return ctx.ReplyError(fmt.Sprintf("コマンド `%s` は存在しません。", cmdName), true)
	}
	
	embed := embed.New().
		SetTitle(fmt.Sprintf("📖 %s コマンド", cmd.Name)).
		SetDescription(cmd.Description).
		SetColor(embed.ColorInfo)
	
	embed.AddField(embed.Field{
		Name:   "使用方法",
		Value:  fmt.Sprintf("`%s%s`", ctx.Bot.GetPrefix(), cmd.Usage),
		Inline: false,
	})
	
	if len(cmd.Aliases) > 0 {
		aliases := make([]string, len(cmd.Aliases))
		for i, alias := range cmd.Aliases {
			aliases[i] = fmt.Sprintf("`%s`", alias)
		}
		embed.AddField(embed.Field{
			Name:   "エイリアス",
			Value:  strings.Join(aliases, ", "),
			Inline: true,
		})
	}
	
	embed.AddField(embed.Field{
		Name:   "カテゴリ",
		Value:  cmd.Category,
		Inline: true,
	})
	
	return ctx.ReplyEmbed(embed.Build(), true)
}

func (r *Registry) statsCommand(ctx *Context) error {
	var m runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m)
	
	uptime := time.Since(startTime)
	
	embed := embed.New().
		SetTitle("📊 Bot 統計情報").
		SetColor(embed.ColorInfo).
		SetThumbnail(embed.Thumbnail{
			URL: ctx.Session.State.User.AvatarURL("256"),
		})
	
	embed.AddField(embed.Field{
		Name:   "稼働時間",
		Value:  format.FormatDuration(uptime),
		Inline: true,
	})
	
	embed.AddField(embed.Field{
		Name:   "メモリ使用量",
		Value:  fmt.Sprintf("%.2f MB", float64(m.Alloc)/1024/1024),
		Inline: true,
	})
	
	embed.AddField(embed.Field{
		Name:   "Goroutines",
		Value:  fmt.Sprintf("%d", runtime.NumGoroutine()),
		Inline: true,
	})
	
	embed.AddField(embed.Field{
		Name:   "サーバー数",
		Value:  fmt.Sprintf("%d", len(ctx.Session.State.Guilds)),
		Inline: true,
	})
	
	embed.AddField(embed.Field{
		Name:   "コマンド数",
		Value:  fmt.Sprintf("%d", r.Count()),
		Inline: true,
	})
	
	embed.AddField(embed.Field{
		Name:   "WebSocket レイテンシ",
		Value:  fmt.Sprintf("%dms", ctx.Session.HeartbeatLatency().Milliseconds()),
		Inline: true,
	})
	
	embed.SetFooter(embed.Footer{
		Text: fmt.Sprintf("Go %s", runtime.Version()),
	})
	
	return ctx.ReplyEmbed(embed.Build())
}

func (r *Registry) statsSlashCommand(ctx *SlashContext) error {
	var m runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m)
	
	uptime := time.Since(startTime)
	
	embed := embed.New().
		SetTitle("📊 Bot 統計情報").
		SetColor(embed.ColorInfo).
		SetThumbnail(embed.Thumbnail{
			URL: ctx.Session.State.User.AvatarURL("256"),
		})
	
	embed.AddField(embed.Field{
		Name:   "稼働時間",
		Value:  format.FormatDuration(uptime),
		Inline: true,
	})
	
	embed.AddField(embed.Field{
		Name:   "メモリ使用量",
		Value:  fmt.Sprintf("%.2f MB", float64(m.Alloc)/1024/1024),
		Inline: true,
	})
	
	embed.AddField(embed.Field{
		Name:   "Goroutines",
		Value:  fmt.Sprintf("%d", runtime.NumGoroutine()),
		Inline: true,
	})
	
	embed.AddField(embed.Field{
		Name:   "サーバー数",
		Value:  fmt.Sprintf("%d", len(ctx.Session.State.Guilds)),
		Inline: true,
	})
	
	embed.AddField(embed.Field{
		Name:   "コマンド数",
		Value:  fmt.Sprintf("%d", r.Count()),
		Inline: true,
	})
	
	embed.AddField(embed.Field{
		Name:   "WebSocket レイテンシ",
		Value:  fmt.Sprintf("%dms", ctx.Session.HeartbeatLatency().Milliseconds()),
		Inline: true,
	})
	
	embed.SetFooter(embed.Footer{
		Text: fmt.Sprintf("Go %s", runtime.Version()),
	})
	
	return ctx.ReplyEmbed(embed.Build(), false)
}

var startTime = time.Now()