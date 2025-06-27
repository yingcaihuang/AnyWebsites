package services

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"anywebsites/internal/config"
)

// ConfigReloadService 配置热重载服务
type ConfigReloadService struct {
	settingsService *SettingsService
	currentConfig   *config.Config
	configMutex     sync.RWMutex
	reloadHandlers  map[string]ReloadHandler
	handlerMutex    sync.RWMutex
	ctx             context.Context
	cancel          context.CancelFunc
}

// ReloadHandler 重载处理器接口
type ReloadHandler interface {
	OnConfigReload(oldConfig, newConfig *config.Config) error
	GetName() string
}

// NewConfigReloadService 创建配置热重载服务
func NewConfigReloadService(settingsService *SettingsService, initialConfig *config.Config) *ConfigReloadService {
	ctx, cancel := context.WithCancel(context.Background())

	service := &ConfigReloadService{
		settingsService: settingsService,
		currentConfig:   initialConfig,
		reloadHandlers:  make(map[string]ReloadHandler),
		ctx:             ctx,
		cancel:          cancel,
	}

	return service
}

// RegisterReloadHandler 注册重载处理器
func (s *ConfigReloadService) RegisterReloadHandler(handler ReloadHandler) {
	s.handlerMutex.Lock()
	defer s.handlerMutex.Unlock()

	s.reloadHandlers[handler.GetName()] = handler
	log.Printf("Registered config reload handler: %s", handler.GetName())
}

// UnregisterReloadHandler 注销重载处理器
func (s *ConfigReloadService) UnregisterReloadHandler(name string) {
	s.handlerMutex.Lock()
	defer s.handlerMutex.Unlock()

	delete(s.reloadHandlers, name)
	log.Printf("Unregistered config reload handler: %s", name)
}

// GetCurrentConfig 获取当前配置
func (s *ConfigReloadService) GetCurrentConfig() *config.Config {
	s.configMutex.RLock()
	defer s.configMutex.RUnlock()

	// 返回配置的副本以避免并发修改
	configCopy := *s.currentConfig
	return &configCopy
}

// ReloadConfig 重载配置
func (s *ConfigReloadService) ReloadConfig() error {
	// 从设置服务构建新配置
	newConfig, err := s.buildConfigFromSettings()
	if err != nil {
		return fmt.Errorf("failed to build config from settings: %w", err)
	}

	// 获取当前配置的副本
	s.configMutex.RLock()
	oldConfig := *s.currentConfig
	s.configMutex.RUnlock()

	// 通知所有处理器
	s.handlerMutex.RLock()
	handlers := make([]ReloadHandler, 0, len(s.reloadHandlers))
	for _, handler := range s.reloadHandlers {
		handlers = append(handlers, handler)
	}
	s.handlerMutex.RUnlock()

	// 执行重载处理器
	var reloadErrors []error
	for _, handler := range handlers {
		if err := handler.OnConfigReload(&oldConfig, newConfig); err != nil {
			log.Printf("Config reload handler %s failed: %v", handler.GetName(), err)
			reloadErrors = append(reloadErrors, fmt.Errorf("handler %s: %w", handler.GetName(), err))
		} else {
			log.Printf("Config reload handler %s succeeded", handler.GetName())
		}
	}

	// 如果有处理器失败，返回错误但仍更新配置
	if len(reloadErrors) > 0 {
		log.Printf("Some config reload handlers failed, but config will still be updated")
	}

	// 更新当前配置
	s.configMutex.Lock()
	s.currentConfig = newConfig
	s.configMutex.Unlock()

	log.Printf("Configuration reloaded successfully")

	// 如果有错误，返回合并的错误信息
	if len(reloadErrors) > 0 {
		return fmt.Errorf("config reloaded with %d handler errors", len(reloadErrors))
	}

	return nil
}

// buildConfigFromSettings 从设置服务构建配置
func (s *ConfigReloadService) buildConfigFromSettings() (*config.Config, error) {
	// 创建新的配置实例
	newConfig := &config.Config{}

	// 服务器配置
	newConfig.Server.Host = s.settingsService.GetStringValue("server", "host", "0.0.0.0")
	newConfig.Server.Port = s.settingsService.GetStringValue("server", "port", "8080")

	// 数据库配置
	newConfig.Database.Host = s.settingsService.GetStringValue("database", "host", "localhost")
	newConfig.Database.Port = s.settingsService.GetStringValue("database", "port", "5432")
	newConfig.Database.User = s.settingsService.GetStringValue("database", "user", "anywebsites")
	newConfig.Database.Password = s.settingsService.GetStringValue("database", "password", "password")
	newConfig.Database.Name = s.settingsService.GetStringValue("database", "name", "anywebsites")
	newConfig.Database.SSLMode = s.settingsService.GetStringValue("database", "ssl_mode", "disable")

	// 上传配置
	newConfig.Upload.MaxFileSize = int64(s.settingsService.GetIntValue("upload", "max_file_size", 10*1024*1024)) // 10MB
	newConfig.Upload.CleanupInterval = s.settingsService.GetIntValue("upload", "cleanup_interval", 3600)
	uploadPath := s.settingsService.GetStringValue("upload", "path", "./uploads")
	newConfig.Upload.Path = uploadPath

	// JWT配置
	newConfig.JWT.Secret = s.settingsService.GetStringValue("security", "jwt_secret", "your-super-secret-jwt-key")

	// 限流配置
	newConfig.RateLimit.Requests = s.settingsService.GetIntValue("security", "rate_limit_requests", 100)
	newConfig.RateLimit.Window = s.settingsService.GetIntValue("security", "rate_limit_window", 3600)

	// GeoIP配置
	newConfig.GeoIP.DBPath = s.settingsService.GetStringValue("geoip", "database_path", "./data/GeoLite2-City.mmdb")

	return newConfig, nil
}

// StartWatching 开始监听配置变化
func (s *ConfigReloadService) StartWatching(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		log.Printf("Config reload service started, checking every %v", interval)

		for {
			select {
			case <-s.ctx.Done():
				log.Println("Config reload service stopped")
				return
			case <-ticker.C:
				// 定期检查配置是否需要重载
				// 这里可以添加更智能的检查逻辑，比如检查设置的更新时间
				if s.shouldReload() {
					if err := s.ReloadConfig(); err != nil {
						log.Printf("Auto config reload failed: %v", err)
					}
				}
			}
		}
	}()
}

// shouldReload 检查是否应该重载配置
func (s *ConfigReloadService) shouldReload() bool {
	// 这里可以实现更复杂的逻辑，比如检查设置的最后更新时间
	// 目前简单返回false，依赖手动触发重载
	return false
}

// Stop 停止配置重载服务
func (s *ConfigReloadService) Stop() {
	s.cancel()
}

// TriggerReload 手动触发配置重载
func (s *ConfigReloadService) TriggerReload() error {
	log.Println("Manual config reload triggered")
	return s.ReloadConfig()
}

// GetReloadHandlers 获取所有注册的重载处理器
func (s *ConfigReloadService) GetReloadHandlers() []string {
	s.handlerMutex.RLock()
	defer s.handlerMutex.RUnlock()

	handlers := make([]string, 0, len(s.reloadHandlers))
	for name := range s.reloadHandlers {
		handlers = append(handlers, name)
	}

	return handlers
}

// ServerReloadHandler 服务器配置重载处理器
type ServerReloadHandler struct {
	name                string
	restartRequiredFile string
}

// NewServerReloadHandler 创建服务器重载处理器
func NewServerReloadHandler() *ServerReloadHandler {
	return &ServerReloadHandler{
		name:                "server",
		restartRequiredFile: ".server_restart_required",
	}
}

// OnConfigReload 处理服务器配置重载
func (h *ServerReloadHandler) OnConfigReload(oldConfig, newConfig *config.Config) error {
	// 检查服务器配置是否有变化
	if oldConfig.Server.Host != newConfig.Server.Host ||
		oldConfig.Server.Port != newConfig.Server.Port {

		log.Printf("Server config changed: %+v -> %+v", oldConfig.Server, newConfig.Server)

		// 检查是否需要重启服务器（端口或主机变化）
		needRestart := false
		restartReason := ""

		if oldConfig.Server.Port != newConfig.Server.Port {
			log.Printf("🔄 Server port changed from %s to %s", oldConfig.Server.Port, newConfig.Server.Port)
			needRestart = true
			restartReason += fmt.Sprintf("Port: %s -> %s; ", oldConfig.Server.Port, newConfig.Server.Port)
		}

		if oldConfig.Server.Host != newConfig.Server.Host {
			log.Printf("🔄 Server host changed from %s to %s", oldConfig.Server.Host, newConfig.Server.Host)
			needRestart = true
			restartReason += fmt.Sprintf("Host: %s -> %s; ", oldConfig.Server.Host, newConfig.Server.Host)
		}

		// 如果需要重启，创建重启标记文件并提供明确指导
		if needRestart {
			newAddr := newConfig.Server.Host + ":" + newConfig.Server.Port

			// 创建重启标记文件
			if err := h.createRestartMarker(newAddr, restartReason); err != nil {
				log.Printf("Failed to create restart marker: %v", err)
			}

			// 提供明确的重启指导
			log.Printf("⚠️  IMPORTANT: Server restart required!")
			log.Printf("📍 Current server address: %s:%s", oldConfig.Server.Host, oldConfig.Server.Port)
			log.Printf("🎯 New server address: %s", newAddr)
			log.Printf("🔧 Changes: %s", restartReason)
			log.Printf("🚀 To apply changes, please restart the server:")
			log.Printf("   1. Stop the current server (Ctrl+C)")
			log.Printf("   2. Restart with: go run cmd/server/main.go")
			log.Printf("   3. Server will start on the new address: %s", newAddr)
			log.Printf("💡 The new configuration has been saved and will be used on next startup.")
		}
	}

	return nil
}

// createRestartMarker 创建重启标记文件
func (h *ServerReloadHandler) createRestartMarker(newAddr, reason string) error {
	content := fmt.Sprintf("RESTART_REQUIRED=true\nNEW_ADDRESS=%s\nREASON=%s\nTIMESTAMP=%d\n",
		newAddr, reason, time.Now().Unix())

	return os.WriteFile(h.restartRequiredFile, []byte(content), 0644)
}

// GetName 获取处理器名称
func (h *ServerReloadHandler) GetName() string {
	return h.name
}

// DatabaseReloadHandler 数据库配置重载处理器
type DatabaseReloadHandler struct {
	name string
}

// NewDatabaseReloadHandler 创建数据库重载处理器
func NewDatabaseReloadHandler() *DatabaseReloadHandler {
	return &DatabaseReloadHandler{
		name: "database",
	}
}

// OnConfigReload 处理数据库配置重载
func (h *DatabaseReloadHandler) OnConfigReload(oldConfig, newConfig *config.Config) error {
	// 检查数据库配置是否有变化
	if oldConfig.Database != newConfig.Database {
		log.Printf("Database config changed: %+v -> %+v", oldConfig.Database, newConfig.Database)

		// 数据库配置变化通常需要重新连接，这里只是记录日志
		log.Printf("WARNING: Database config changed, connection pool restart may be required")
	}

	return nil
}

// GetName 获取处理器名称
func (h *DatabaseReloadHandler) GetName() string {
	return h.name
}

// SecurityReloadHandler 安全配置重载处理器
type SecurityReloadHandler struct {
	name string
}

// NewSecurityReloadHandler 创建安全重载处理器
func NewSecurityReloadHandler() *SecurityReloadHandler {
	return &SecurityReloadHandler{
		name: "security",
	}
}

// OnConfigReload 处理安全配置重载
func (h *SecurityReloadHandler) OnConfigReload(oldConfig, newConfig *config.Config) error {
	// 检查JWT配置是否有变化
	if oldConfig.JWT.Secret != newConfig.JWT.Secret {
		log.Printf("JWT secret changed")
		// JWT密钥变化会影响现有token的验证
		log.Printf("WARNING: JWT secret changed, existing tokens may become invalid")
	}

	// 检查限流配置是否有变化
	if oldConfig.RateLimit.Requests != newConfig.RateLimit.Requests ||
		oldConfig.RateLimit.Window != newConfig.RateLimit.Window {
		log.Printf("Rate limit config changed: %+v -> %+v",
			oldConfig.RateLimit, newConfig.RateLimit)
	}

	return nil
}

// GetName 获取处理器名称
func (h *SecurityReloadHandler) GetName() string {
	return h.name
}
