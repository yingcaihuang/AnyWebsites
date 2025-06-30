package database

import (
	"fmt"
	"log"

	"anywebsites/internal/auth"
	"anywebsites/internal/config"
	"anywebsites/internal/models"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// Connect 连接数据库
func Connect(cfg *config.Config) error {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Name,
		cfg.Database.SSLMode,
	)

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	log.Println("Database connected successfully")
	return nil
}

// Migrate 执行数据库迁移
func Migrate() error {
	if DB == nil {
		return fmt.Errorf("database connection is nil")
	}

	// 启用 UUID 扩展
	if err := DB.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"").Error; err != nil {
		log.Printf("Warning: failed to create uuid-ossp extension: %v", err)
	}

	log.Println("Starting database initialization")

	// 完全跳过GORM自动迁移，使用SQL脚本初始化
	log.Println("Using SQL script for database initialization")
	log.Println("All tables will be created from database-init.sql")

	log.Println("Database initialization completed successfully")

	// 初始化系统设置
	if err := InitializeSystemSettings(); err != nil {
		log.Printf("Warning: failed to initialize system settings: %v", err)
	}

	// 初始化默认管理员账户
	if err := InitializeDefaultAdmins(); err != nil {
		log.Printf("Warning: failed to initialize default admins: %v", err)
	}

	return nil
}

// CreateIndexes 创建数据库索引
func CreateIndexes() error {
	if DB == nil {
		return fmt.Errorf("database connection is nil")
	}

	// 为各表创建复合索引
	indexes := []string{
		// ContentAnalytics 索引
		"CREATE INDEX IF NOT EXISTS idx_content_analytics_content_time ON content_analytics(content_id, access_time)",
		"CREATE INDEX IF NOT EXISTS idx_content_analytics_user_time ON content_analytics(user_id, access_time)",
		"CREATE INDEX IF NOT EXISTS idx_content_analytics_ip_time ON content_analytics(ip_address, access_time)",
		"CREATE INDEX IF NOT EXISTS idx_content_analytics_country ON content_analytics(country)",
		"CREATE INDEX IF NOT EXISTS idx_contents_user_active ON contents(user_id, is_active)",
		"CREATE INDEX IF NOT EXISTS idx_contents_expires_at ON contents(expires_at)",

		// SystemSettings 索引
		"CREATE INDEX IF NOT EXISTS idx_system_settings_category_key ON system_settings(category, key)",
		"CREATE INDEX IF NOT EXISTS idx_system_settings_active ON system_settings(is_active)",
		"CREATE INDEX IF NOT EXISTS idx_system_settings_category ON system_settings(category)",
		"CREATE INDEX IF NOT EXISTS idx_system_setting_histories_setting_id ON system_setting_histories(setting_id)",
		"CREATE INDEX IF NOT EXISTS idx_system_setting_histories_created_at ON system_setting_histories(created_at)",
		"CREATE INDEX IF NOT EXISTS idx_system_setting_categories_sort_order ON system_setting_categories(sort_order)",
	}

	for _, index := range indexes {
		if err := DB.Exec(index).Error; err != nil {
			log.Printf("Warning: failed to create index: %v", err)
		}
	}

	log.Println("Database indexes created successfully")
	return nil
}

// Close 关闭数据库连接
func Close() error {
	if DB == nil {
		return nil
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}

	return sqlDB.Close()
}

// InitializeSystemSettings 初始化系统设置
func InitializeSystemSettings() error {
	if DB == nil {
		return fmt.Errorf("database connection is nil")
	}

	// 创建默认设置分类
	categories := []models.SystemSettingCategory{
		{
			Name:        "server",
			DisplayName: "服务器设置",
			Description: "服务器相关配置，包括端口、主机等",
			Icon:        "bi-server",
			SortOrder:   1,
		},
		{
			Name:        "database",
			DisplayName: "数据库设置",
			Description: "数据库连接和配置参数",
			Icon:        "bi-database",
			SortOrder:   2,
		},
		{
			Name:        "upload",
			DisplayName: "上传设置",
			Description: "文件上传相关配置",
			Icon:        "bi-cloud-upload",
			SortOrder:   3,
		},
		{
			Name:        "security",
			DisplayName: "安全设置",
			Description: "安全相关配置，包括JWT、限流等",
			Icon:        "bi-shield-check",
			SortOrder:   4,
		},
		{
			Name:        "geoip",
			DisplayName: "地理位置设置",
			Description: "GeoIP服务相关配置",
			Icon:        "bi-globe",
			SortOrder:   5,
		},
		{
			Name:        "system",
			DisplayName: "系统设置",
			Description: "系统级别的配置参数",
			Icon:        "bi-gear",
			SortOrder:   6,
		},
	}

	// 创建分类（如果不存在）
	for _, category := range categories {
		var existingCategory models.SystemSettingCategory
		if err := DB.Where("name = ?", category.Name).First(&existingCategory).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				if err := DB.Create(&category).Error; err != nil {
					log.Printf("Failed to create category %s: %v", category.Name, err)
				}
			}
		}
	}

	log.Println("System settings initialized successfully")
	return nil
}

// InitializeDefaultAdmins 初始化默认管理员账户
func InitializeDefaultAdmins() error {
	if DB == nil {
		return fmt.Errorf("database connection is nil")
	}

	// 默认管理员账户配置
	defaultAdmins := []struct {
		Username string
		Email    string
		Password string
	}{
		{
			Username: "admin",
			Email:    "admin@example.com",
			Password: "Google@google",
		},
		{
			Username: "yingcai",
			Email:    "yingcai@yingcai.com",
			Password: "Yingcai@yingcai",
		},
	}

	for _, admin := range defaultAdmins {
		// 检查用户是否已存在
		var existingUser models.User
		if err := DB.Where("username = ? OR email = ?", admin.Username, admin.Email).First(&existingUser).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				// 用户不存在，创建新用户
				hashedPassword, err := hashPassword(admin.Password)
				if err != nil {
					log.Printf("Failed to hash password for user %s: %v", admin.Username, err)
					continue
				}

				// 生成API密钥
				apiKey, err := generateAPIKey()
				if err != nil {
					log.Printf("Failed to generate API key for user %s: %v", admin.Username, err)
					continue
				}

				newUser := models.User{
					Username: admin.Username,
					Email:    admin.Email,
					Password: hashedPassword,
					APIKey:   apiKey,
					IsActive: true,
					IsAdmin:  true,
				}

				if err := DB.Create(&newUser).Error; err != nil {
					log.Printf("Failed to create default admin user %s: %v", admin.Username, err)
				} else {
					log.Printf("Created default admin user: %s", admin.Username)
				}
			} else {
				log.Printf("Error checking for existing user %s: %v", admin.Username, err)
			}
		} else {
			// 用户已存在，更新密码为默认密码
			hashedPassword, err := hashPassword(admin.Password)
			if err != nil {
				log.Printf("Failed to hash password for existing user %s: %v", admin.Username, err)
				continue
			}

			// 更新用户密码和管理员状态
			if err := DB.Model(&existingUser).Updates(map[string]interface{}{
				"password":  hashedPassword,
				"is_admin":  true,
				"is_active": true,
			}).Error; err != nil {
				log.Printf("Failed to update default admin user %s: %v", admin.Username, err)
			} else {
				log.Printf("Updated default admin user %s with new password", admin.Username)
			}
		}
	}

	return nil
}

// hashPassword 对密码进行哈希加密
func hashPassword(password string) (string, error) {
	// 使用 auth 包中的 HashPassword 函数
	return auth.HashPassword(password)
}

// generateAPIKey 生成API密钥
func generateAPIKey() (string, error) {
	// 生成UUID并取前32个字符作为API密钥
	return uuid.New().String()[:32], nil
}
