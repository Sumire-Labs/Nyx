package commands

import (
	"fmt"
	"time"

	nyxembed "github.com/Sumire-Labs/Nyx-API/embed"
	"github.com/bwmarrin/discordgo"
)

func init() {
	pingCommand := &Command{
		Name:        "ping",
		Description: "Botのレイテンシを確認します",
		Usage:       "ping",
		Aliases:     []string{"pong", "latency"},
		Category:    "General",
		GuildOnly:   false,
		OwnerOnly:   false,
		SlashCommand: &discordgo.ApplicationCommand{
			Name:        "ping",
			Description: "Botのレイテンシを確認します",
		},
		Execute:      executePing,
		ExecuteSlash: executePingSlash,
	}

	DefaultCommands = append(DefaultCommands, pingCommand)
}

func executePing(ctx *Context) error {
	startTime := time.Now()
	
	message, err := ctx.Session.ChannelMessageSend(ctx.Message.ChannelID, "🏓 Pinging...")
	if err != nil {
		return err
	}

	wsLatency := ctx.Session.HeartbeatLatency()
	apiLatency := time.Since(startTime)

	embed := nyxembed.Info("🏓 Pong!", "").
		AddField(nyxembed.Field{
			Name:   "APIレイテンシ",
			Value:  fmt.Sprintf("%dms", apiLatency.Milliseconds()),
			Inline: true,
		}).
		AddField(nyxembed.Field{
			Name:   "WebSocketレイテンシ",
			Value:  fmt.Sprintf("%dms", wsLatency.Milliseconds()),
			Inline: true,
		}).
		SetTimestamp(time.Now()).
		Build()

	_, err = ctx.Session.ChannelMessageEditEmbed(ctx.Message.ChannelID, message.ID, embed)
	return err
}

func executePingSlash(ctx *SlashContext) error {
	startTime := time.Now()

	err := ctx.Reply("🏓 Pinging...", false)
	if err != nil {
		return err
	}

	wsLatency := ctx.Session.HeartbeatLatency()
	apiLatency := time.Since(startTime)

	embed := nyxembed.Info("🏓 Pong!", "").
		AddField(nyxembed.Field{
			Name:   "APIレイテンシ",
			Value:  fmt.Sprintf("%dms", apiLatency.Milliseconds()),
			Inline: true,
		}).
		AddField(nyxembed.Field{
			Name:   "WebSocketレイテンシ",
			Value:  fmt.Sprintf("%dms", wsLatency.Milliseconds()),
			Inline: true,
		}).
		SetTimestamp(time.Now()).
		Build()

	return ctx.Session.InteractionResponseEdit(ctx.Interaction, &discordgo.WebhookEdit{
		Embeds: &[]*discordgo.MessageEmbed{embed},
	})
}