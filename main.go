package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Sumire-Labs/Nyx-API/logger"
	"github.com/Sumire-Labs/Nyx/bot"
	"github.com/Sumire-Labs/Nyx/commands"
	"github.com/Sumire-Labs/Nyx/database"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Bot struct {
		Token    string `yaml:"token"`
		Prefix   string `yaml:"prefix"`
		Status   string `yaml:"status"`
		Activity struct {
			Type string `yaml:"type"`
			Name string `yaml:"name"`
		} `yaml:"activity"`
	} `yaml:"bot"`
	
	Database struct {
		Path string `yaml:"path"`
	} `yaml:"database"`
	
	Logging struct {
		Level      string `yaml:"level"`
		File       string `yaml:"file"`
		WebhookURL string `yaml:"webhook_url"`
	} `yaml:"logging"`
	
	Features struct {
		SlashCommands   bool   `yaml:"slash_commands"`
		AutoRole        bool   `yaml:"auto_role"`
		WelcomeMessage  bool   `yaml:"welcome_message"`
		LoggingChannel  string `yaml:"logging_channel"`
	} `yaml:"features"`
	
	OwnerIDs []string `yaml:"owner_ids"`
}

var (
	configPath = flag.String("config", "config.yaml", "Path to config file")
	log        *logger.Logger
)

func main() {
	flag.Parse()
	
	config, err := loadConfig(*configPath)
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}
	
	logLevel := parseLogLevel(config.Logging.Level)
	log = logger.New("Nyx", logLevel)
	
	if config.Logging.File != "" {
		if err := log.SetFileOutput(config.Logging.File); err != nil {
			log.Warn("Failed to set file output: %v", err)
		}
	}
	
	if config.Logging.WebhookURL != "" {
		log.SetWebhook(config.Logging.WebhookURL, "Nyx Bot", "")
	}
	
	log.Info("Starting Nyx Bot...")
	
	db, err := database.New(config.Database.Path)
	if err != nil {
		log.Fatal("Failed to initialize database: %v", err)
	}
	defer db.Close()
	
	log.Info("Database initialized")
	
	commandRegistry := commands.NewRegistry()
	commandRegistry.RegisterDefaultCommands()
	log.Info("Registered %d commands", commandRegistry.Count())
	
	botInstance, err := bot.New(bot.Config{
		Token:          config.Bot.Token,
		Prefix:         config.Bot.Prefix,
		Status:         config.Bot.Status,
		ActivityType:   config.Bot.Activity.Type,
		ActivityName:   config.Bot.Activity.Name,
		Database:       db,
		Commands:       commandRegistry,
		Logger:         log,
		SlashCommands:  config.Features.SlashCommands,
		LoggingChannel: config.Features.LoggingChannel,
		OwnerIDs:       config.OwnerIDs,
	})
	
	if err != nil {
		log.Fatal("Failed to create bot: %v", err)
	}
	
	if err := botInstance.Start(); err != nil {
		log.Fatal("Failed to start bot: %v", err)
	}
	
	log.Info("Bot is now running. Press CTRL+C to exit.")
	
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
	
	log.Info("Shutting down bot...")
	if err := botInstance.Stop(); err != nil {
		log.Error("Error during shutdown: %v", err)
	}
	
	log.Info("Bot shut down successfully")
}

func loadConfig(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	
	var config Config
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, err
	}
	
	if config.Bot.Token == "" {
		return nil, fmt.Errorf("bot token is required")
	}
	
	if config.Bot.Prefix == "" {
		config.Bot.Prefix = "!"
	}
	
	return &config, nil
}

func parseLogLevel(level string) logger.LogLevel {
	switch level {
	case "debug":
		return logger.DebugLevel
	case "info":
		return logger.InfoLevel
	case "warn":
		return logger.WarnLevel
	case "error":
		return logger.ErrorLevel
	default:
		return logger.InfoLevel
	}
}