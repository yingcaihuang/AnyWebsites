package main

import (
	"fmt"
	"log"

	"anywebsites/internal/api"
	"anywebsites/internal/config"
	"anywebsites/internal/database"
	"anywebsites/internal/services"
)

func main() {
	// 加载配置
	cfg := config.Load()

	// 连接数据库
	if err := database.Connect(cfg); err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// 执行数据库迁移
	if err := database.Migrate(); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// 初始化系统设置服务
	settingsService := services.NewSettingsService()

	// 初始化 GeoIP 服务
	geoipPath := "data/geoip/GeoLite2-City.mmdb"
	geoipService, err := services.NewGeoIPService(geoipPath)
	if err != nil {
		log.Printf("Warning: Failed to initialize GeoIP service: %v", err)
		log.Printf("Geographic location features will be limited")
		geoipService = nil
	} else {
		log.Printf("GeoIP service initialized successfully")
		defer geoipService.Close()
	}

	// 启动清理服务
	cleanupService := services.NewCleanupService()
	go cleanupService.Start()
	defer cleanupService.Stop()

	// 设置路由
	r := api.SetupRoutes(cfg, geoipService)

	// 从数据库加载服务器配置
	serverHost := cfg.Server.Host // 默认使用静态配置的主机
	serverPort := cfg.Server.Port // 默认使用静态配置的端口

	// 尝试从数据库获取服务器配置
	if hostSetting, err := settingsService.GetSetting("server", "host"); err == nil && hostSetting != nil {
		host := hostSetting.GetStringValue()
		if host != "" {
			serverHost = host
		}
	}

	if portSetting, err := settingsService.GetSetting("server", "port"); err == nil && portSetting != nil {
		if port, err := portSetting.GetIntValue(); err == nil && port > 0 {
			serverPort = fmt.Sprintf("%d", port)
		}
	}

	// 启动服务器
	addr := serverHost + ":" + serverPort
	log.Printf("Server starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
