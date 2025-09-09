package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Sumire-Labs/Nyx-API/logger"
	"github.com/Sumire-Labs/Nyx-API/utils"
	"github.com/Sumire-Labs/Nyx/bot"
	"github.com/Sumire-Labs/Nyx/commands"
	"github.com/Sumire-Labs/Nyx/database"
	"github.com/Sumire-Labs/Nyx/di"
	nyxutils "github.com/Sumire-Labs/Nyx/utils"
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
		SlashCommands  bool   `yaml:"slash_commands"`
		AutoRole       bool   `yaml:"auto_role"`
		WelcomeMessage bool   `yaml:"welcome_message"`
		LoggingChannel string `yaml:"logging_channel"`
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

	// 🚀 DIコンテナを初期化
	container := di.NewContainer()

	// DIコンテナ用の設定を変換
	diConfig := &di.Config{
		Bot: struct {
			Token        string
			Prefix       string
			Status       string
			ActivityType string
			ActivityName string
			OwnerIDs     []string
		}{
			Token:        config.Bot.Token,
			Prefix:       config.Bot.Prefix,
			Status:       config.Bot.Status,
			ActivityType: config.Bot.Activity.Type,
			ActivityName: config.Bot.Activity.Name,
			OwnerIDs:     config.OwnerIDs,
		},
		Database: struct {
			Path string
		}{
			Path: config.Database.Path,
		},
		Logging: struct {
			Level      string
			File       string
			WebhookURL string
		}{
			Level:      config.Logging.Level,
			File:       config.Logging.File,
			WebhookURL: config.Logging.WebhookURL,
		},
		Features: struct {
			SlashCommands  bool
			LoggingChannel string
		}{
			SlashCommands:  config.Features.SlashCommands,
			LoggingChannel: config.Features.LoggingChannel,
		},
	}

	// 🔧 サービスを登録
	if err := di.RegisterServices(container, diConfig); err != nil {
		fmt.Printf("Failed to register services: %v\n", err)
		os.Exit(1)
	}

	// 🎯 ServiceLocatorを作成
	serviceLocator := di.NewServiceLocator(container)

	// 🔧 FIXED: エラーハンドリング追加（パニック対策）
	loggerService, err := serviceLocator.Logger()
	if err != nil {
		fmt.Printf("❌ Failed to get logger service: %v\n", err)
		os.Exit(1)
	}
	log = loggerService.(*logger.Logger)
	log.Info("🚀 Starting Nyx Bot with Dependency Injection...")

	dbService, err := serviceLocator.Database()
	if err != nil {
		log.Fatal("❌ Failed to get database service: %v", err)
	}
	db := dbService.(*database.Database)
	defer db.Close()
	log.Info("✅ Database initialized")

	commandService, err := serviceLocator.Commands()
	if err != nil {
		log.Fatal("❌ Failed to get command service: %v", err)
	}
	log.Info("📋 Registered %d commands", commandService.Count())

	// 実際の Registry を取得（Bot設定用）
	commandRegistry, _ := container.Get("Commands")
	registry := commandRegistry.(*commands.Registry)

	// 🤖 Botインスタンスを作成（完全DI対応）
	botInstance, err := bot.New(bot.Config{
		Token:          config.Bot.Token,
		Prefix:         config.Bot.Prefix,
		Status:         config.Bot.Status,
		ActivityType:   config.Bot.Activity.Type,
		ActivityName:   config.Bot.Activity.Name,
		Database:       db,
		Commands:       registry,
		Logger:         log,
		Services:       serviceLocator, // 🎯 DI ServiceLocator設定
		SlashCommands:  config.Features.SlashCommands,
		LoggingChannel: config.Features.LoggingChannel,
		OwnerIDs:       config.OwnerIDs,
	})

	if err != nil {
		log.Fatal("❌ Failed to create bot: %v", err)
	}

	if err := botInstance.Start(); err != nil {
		log.Fatal("❌ Failed to start bot: %v", err)
	}

	log.Info("✅ Nyx Bot is now running with DI! Press CTRL+C to exit.")

	// Nyx-API 0.3.1で実装された実験的なグローバルリソースクリーンアップを実装。
	defer func() {
		log.Info("🧹 Cleaning up global resources...")
		utils.CleanupGlobalRateLimiter()
		utils.CleanupGlobalDiscordCache()
		log.Info("✨ Global resources cleaned up")
	}()

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	log.Info("🛑 Shutting down bot...")
	if err := botInstance.Stop(); err != nil {
		log.Error("❌ Error during shutdown: %v", err)
	}

	log.Info("✅ Bot shut down successfully")
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

	// 🔧 FIXED: Botトークンのセキュリティ検証を追加
	if config.Bot.Token == "" {
		return nil, fmt.Errorf("bot token is required")
	}
	
	if err := validateBotToken(config.Bot.Token); err != nil {
		return nil, fmt.Errorf("invalid bot token: %w", err)
	}

	// 🔧 FIXED: Owner ID検証を追加
	for i, ownerID := range config.OwnerIDs {
		if ownerID == "" {
			continue // 空のIDはスキップ
		}
		if !nyxutils.ValidateDiscordID(ownerID) {
			return nil, fmt.Errorf("invalid owner ID at position %d: %s", i, ownerID)
		}
	}

	if config.Bot.Prefix == "" {
		config.Bot.Prefix = "!"
	}

	// 🔧 FIXED: プレフィックス検証を追加
	config.Bot.Prefix = nyxutils.SanitizeInput(config.Bot.Prefix)
	if len(config.Bot.Prefix) > 10 {
		return nil, fmt.Errorf("bot prefix too long (max 10 characters)")
	}

	return &config, nil
}

// 🔧 FIXED: Botトークン検証関数を追加
func validateBotToken(token string) error {
	// Discordボットトークンの基本形式を検証
	// MTxxxxxxxxxxxxxxxxxxxxxx.xxxxxx.xxxxxxxxxxxxxxxxxxxxxxxxxxx の形式
	
	if len(token) < 50 {
		return fmt.Errorf("token too short (minimum 50 characters)")
	}
	
	if len(token) > 100 {
		return fmt.Errorf("token too long (maximum 100 characters)")
	}
	
	// トークンにスペースや改行が含まれていないかチェック
	sanitized := nyxutils.SanitizeInput(token)
	if sanitized != token {
		return fmt.Errorf("token contains invalid characters")
	}
	
	// 悪意のあるコンテンツパターンをチェック
	if nyxutils.ContainsMaliciousContent(token) {
		return fmt.Errorf("token contains potentially malicious content")
	}
	
	// Discordボットトークンの基本パターンを検証 (最低限の形式チェック)
	// 注意: 実際のトークン形式は変更される可能性があるため、基本的なチェックのみ
	if token == "your_bot_token_here" || token == "BOT_TOKEN" || token == "token" {
		return fmt.Errorf("placeholder token detected, please use actual bot token")
	}
	
	return nil
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
