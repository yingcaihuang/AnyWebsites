package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config 应用配置结构
type Config struct {
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	Server   ServerConfig
	Upload   UploadConfig
	GeoIP    GeoIPConfig
	Admin    AdminConfig
	RateLimit RateLimitConfig
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

// RedisConfig Redis 配置
type RedisConfig struct {
	Host     string
	Port     string
	Password string
}

// JWTConfig JWT 配置
type JWTConfig struct {
	Secret string
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Host string
	Port string
}

// UploadConfig 上传配置
type UploadConfig struct {
	Path            string
	MaxFileSize     int64
	CleanupInterval int
}

// GeoIPConfig GeoIP 配置
type GeoIPConfig struct {
	DBPath string
}

// AdminConfig 管理员配置
type AdminConfig struct {
	Username string
	Password string
}

// RateLimitConfig 限流配置
type RateLimitConfig struct {
	Requests int
	Window   int
}

// Load 加载配置
func Load() *Config {
	// 加载 .env 文件
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	return &Config{
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "anywebsites"),
			Password: getEnv("DB_PASSWORD", "password"),
			Name:     getEnv("DB_NAME", "anywebsites"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
		},
		JWT: JWTConfig{
			Secret: getEnv("JWT_SECRET", "your-super-secret-jwt-key"),
		},
		Server: ServerConfig{
			Host: getEnv("SERVER_HOST", "0.0.0.0"),
			Port: getEnv("SERVER_PORT", "8080"),
		},
		Upload: UploadConfig{
			Path:            getEnv("UPLOAD_PATH", "./uploads"),
			MaxFileSize:     getEnvAsInt64("MAX_FILE_SIZE", 10485760), // 10MB
			CleanupInterval: getEnvAsInt("CLEANUP_INTERVAL", 3600),    // 1 hour
		},
		GeoIP: GeoIPConfig{
			DBPath: getEnv("GEOIP_DB_PATH", "./data/GeoLite2-City.mmdb"),
		},
		Admin: AdminConfig{
			Username: getEnv("ADMIN_USERNAME", "admin"),
			Password: getEnv("ADMIN_PASSWORD", "admin123"),
		},
		RateLimit: RateLimitConfig{
			Requests: getEnvAsInt("RATE_LIMIT_REQUESTS", 100),
			Window:   getEnvAsInt("RATE_LIMIT_WINDOW", 3600),
		},
	}
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt 获取环境变量并转换为整数
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvAsInt64 获取环境变量并转换为 int64
func getEnvAsInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intValue
		}
	}
	return defaultValue
}
