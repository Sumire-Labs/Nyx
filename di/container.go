package di

import (
	"fmt"
	"reflect"
	"sync"
)

// Container DIコンテナの実装
type Container struct {
	services  map[string]interface{}
	providers map[string]ServiceProvider
	mu        sync.RWMutex
}

// ServiceProvider サービスプロバイダー関数の型
type ServiceProvider func() (interface{}, error)

// NewContainer 新しいDIコンテナを作成
func NewContainer() *Container {
	return &Container{
		services:  make(map[string]interface{}),
		providers: make(map[string]ServiceProvider),
	}
}

// Register サービスプロバイダーを登録
func (c *Container) Register(name string, provider ServiceProvider) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	if _, exists := c.providers[name]; exists {
		return fmt.Errorf("service %s is already registered", name)
	}
	
	c.providers[name] = provider
	return nil
}

// RegisterInstance サービスインスタンスを直接登録
func (c *Container) RegisterInstance(name string, instance interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	if _, exists := c.services[name]; exists {
		return fmt.Errorf("service instance %s is already registered", name)
	}
	
	c.services[name] = instance
	return nil
}

// Get サービスを名前で取得
func (c *Container) Get(name string) (interface{}, error) {
	c.mu.RLock()
	// 既にインスタンス化されている場合はそれを返す
	if service, exists := c.services[name]; exists {
		c.mu.RUnlock()
		return service, nil
	}
	
	// プロバイダーが存在するか確認
	provider, exists := c.providers[name]
	if !exists {
		c.mu.RUnlock()
		return nil, fmt.Errorf("service %s is not registered", name)
	}
	c.mu.RUnlock()
	
	// サービスを作成してキャッシュ
	service, err := provider()
	if err != nil {
		return nil, fmt.Errorf("failed to create service %s: %w", name, err)
	}
	
	c.mu.Lock()
	c.services[name] = service
	c.mu.Unlock()
	
	return service, nil
}

// GetService サービスを型で取得
func (c *Container) GetService(servicePtr interface{}) error {
	v := reflect.ValueOf(servicePtr)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Interface {
		return fmt.Errorf("servicePtr must be a pointer to an interface")
	}
	
	interfaceType := v.Elem().Type()
	serviceName := getServiceNameFromType(interfaceType)
	
	service, err := c.Get(serviceName)
	if err != nil {
		return err
	}
	
	serviceValue := reflect.ValueOf(service)
	if !serviceValue.Type().Implements(interfaceType) {
		return fmt.Errorf("service does not implement the required interface")
	}
	
	v.Elem().Set(serviceValue)
	return nil
}

// MustGet サービスを取得（エラー時はパニック）
func (c *Container) MustGet(name string) interface{} {
	service, err := c.Get(name)
	if err != nil {
		panic(fmt.Sprintf("failed to get service %s: %v", name, err))
	}
	return service
}

// Clear すべてのサービスをクリア
func (c *Container) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	c.services = make(map[string]interface{})
	c.providers = make(map[string]ServiceProvider)
}

// IsRegistered サービスが登録されているかチェック
func (c *Container) IsRegistered(name string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	_, hasInstance := c.services[name]
	_, hasProvider := c.providers[name]
	
	return hasInstance || hasProvider
}

// GetRegisteredServices 登録されているサービス名のリストを取得
func (c *Container) GetRegisteredServices() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	var services []string
	
	for name := range c.services {
		services = append(services, name)
	}
	
	for name := range c.providers {
		if _, exists := c.services[name]; !exists {
			services = append(services, name)
		}
	}
	
	return services
}

// getServiceNameFromType インターフェース型からサービス名を推定
func getServiceNameFromType(t reflect.Type) string {
	name := t.Name()
	if name == "" {
		name = t.String()
	}
	
	// "Service" サフィックスを除去してサービス名を作成
	if len(name) > 7 && name[len(name)-7:] == "Service" {
		return name[:len(name)-7]
	}
	
	return name
}