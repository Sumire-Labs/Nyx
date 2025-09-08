package di

import (
	"fmt"

	"github.com/Sumire-Labs/Nyx-API/logger"
	"github.com/Sumire-Labs/Nyx/commands"
	"github.com/Sumire-Labs/Nyx/database"
	"github.com/Sumire-Labs/Nyx/services"
)

// Config DIコンテナ用の設定構造体
type Config struct {
	Bot struct {
		Token        string
		Prefix       string
		Status       string
		ActivityType string
		ActivityName string
		OwnerIDs     []string
	}
	Database struct {
		Path string
	}
	Logging struct {
		Level      string
		File       string
		WebhookURL string
	}
	Features struct {
		SlashCommands  bool
		LoggingChannel string
	}
}

// ServiceLocatorImpl ServiceLocator インターフェースの実装
type ServiceLocatorImpl struct {
	container *Container
}

// CommandServiceAdapter commands.Registry を services.CommandService に適応させる
type CommandServiceAdapter struct {
	registry *commands.Registry
}

// GetRegistry 内部の Registry を取得
func (c *CommandServiceAdapter) GetRegistry() *commands.Registry {
	return c.registry
}

func (c *CommandServiceAdapter) Register(cmd interface{}) {
	if command, ok := cmd.(*commands.Command); ok {
		c.registry.Register(command)
	}
}

func (c *CommandServiceAdapter) GetCommand(name string) (interface{}, bool) {
	return c.registry.GetCommand(name)
}

func (c *CommandServiceAdapter) GetAll() map[string]interface{} {
	all := c.registry.GetAll()
	result := make(map[string]interface{})
	for k, v := range all {
		result[k] = interface{}(v)
	}
	return result
}

func (c *CommandServiceAdapter) Count() int {
	return c.registry.Count()
}

func (c *CommandServiceAdapter) GetByCategory(category string) []interface{} {
	cmds := c.registry.GetByCategory(category)
	result := make([]interface{}, len(cmds))
	for i, cmd := range cmds {
		result[i] = cmd
	}
	return result
}

func (c *CommandServiceAdapter) RegisterDefaultCommands() {
	c.registry.RegisterDefaultCommands()
}

// BotConfigInfoImpl BotConfigInfo インターフェースの実装
type BotConfigInfoImpl struct {
	prefix   string
	ownerIDs []string
}

func (b *BotConfigInfoImpl) GetPrefix() string {
	return b.prefix
}

func (b *BotConfigInfoImpl) IsOwner(userID string) bool {
	for _, ownerID := range b.ownerIDs {
		if userID == ownerID {
			return true
		}
	}
	return false
}

// NewServiceLocator ServiceLocator の新しいインスタンスを作成
func NewServiceLocator(container *Container) services.ServiceLocator {
	return &ServiceLocatorImpl{container: container}
}

func (sl *ServiceLocatorImpl) Database() services.DatabaseService {
	service, err := sl.container.Get("Database")
	if err != nil {
		panic(fmt.Sprintf("failed to get database service: %v", err))
	}
	return service.(services.DatabaseService)
}

func (sl *ServiceLocatorImpl) Logger() services.LoggingService {
	service, err := sl.container.Get("Logger")
	if err != nil {
		panic(fmt.Sprintf("failed to get logger service: %v", err))
	}
	return service.(services.LoggingService)
}

func (sl *ServiceLocatorImpl) Commands() services.CommandService {
	service, err := sl.container.Get("Commands")
	if err != nil {
		panic(fmt.Sprintf("failed to get commands service: %v", err))
	}
	// *commands.Registry を services.CommandService として返す
	registry := service.(*commands.Registry)
	return &CommandServiceAdapter{registry: registry}
}

func (sl *ServiceLocatorImpl) BotConfig() services.BotConfigInfo {
	service, err := sl.container.Get("BotConfig")
	if err != nil {
		panic(fmt.Sprintf("failed to get bot config service: %v", err))
	}
	return service.(services.BotConfigInfo)
}

// RegisterServices DIコンテナにすべてのサービスを登録
func RegisterServices(container *Container, config *Config) error {
	// Database Service
	if err := container.Register("Database", func() (interface{}, error) {
		db, err := database.New(config.Database.Path)
		if err != nil {
			return nil, fmt.Errorf("failed to create database: %w", err)
		}
		return db, nil
	}); err != nil {
		return fmt.Errorf("failed to register database service: %w", err)
	}

	// Logger Service
	if err := container.Register("Logger", func() (interface{}, error) {
		logLevel := parseLogLevel(config.Logging.Level)
		log := logger.New("Nyx", logLevel)
		
		if config.Logging.File != "" {
			if err := log.SetFileOutput(config.Logging.File); err != nil {
				log.Warn("Failed to set file output: %v", err)
			}
		}
		
		if config.Logging.WebhookURL != "" {
			log.SetWebhook(config.Logging.WebhookURL, "Nyx Bot", "")
		}
		
		return log, nil
	}); err != nil {
		return fmt.Errorf("failed to register logger service: %w", err)
	}

	// Commands Service
	if err := container.Register("Commands", func() (interface{}, error) {
		registry := commands.NewRegistry()
		registry.RegisterDefaultCommands()
		return registry, nil
	}); err != nil {
		return fmt.Errorf("failed to register commands service: %w", err)
	}

	// BotConfig Service
	if err := container.Register("BotConfig", func() (interface{}, error) {
		return &BotConfigInfoImpl{
			prefix:   config.Bot.Prefix,
			ownerIDs: config.Bot.OwnerIDs,
		}, nil
	}); err != nil {
		return fmt.Errorf("failed to register bot config service: %w", err)
	}
	
	// Bot Service は main.go で直接管理されるため、ここでは登録しない

	return nil
}

// parseLogLevel ログレベル文字列を解析
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

// ConvertMainConfig main.go の Config を DI 用 Config に変換
func ConvertMainConfig(mainConfig interface{}) (*Config, error) {
	// この関数は main.go の Config 構造体と互換性を保つため
	// 実際の実装では reflection を使用するか、
	// main.go の Config を直接使用するように修正する必要があります
	
	// 仮の実装として、interface{} からの変換を試みます
	// 実際の使用時には適切な型アサーションまたは変換ロジックを実装してください
	
	config := &Config{}
	
	// ここで mainConfig からフィールドをコピーする実装が必要
	// 今回は仮実装として空の構造体を返します
	
	return config, nil
}