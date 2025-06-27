package services

import (
	"errors"
	"fmt"
	"log"
	"time"

	"anywebsites/internal/auth"
	"anywebsites/internal/database"
	"anywebsites/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserService 用户服务
type UserService struct {
	db *gorm.DB
}

// NewUserService 创建用户服务实例
func NewUserService(db *gorm.DB) *UserService {
	return &UserService{
		db: db,
	}
}

// RegisterRequest 注册请求结构
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// LoginRequest 登录请求结构
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse 登录响应结构
type LoginResponse struct {
	User         *models.User `json:"user"`
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
}

// RefreshRequest 刷新令牌请求结构
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// Register 用户注册
func (s *UserService) Register(req *RegisterRequest) (*models.User, error) {
	// 检查用户名是否已存在
	var existingUser models.User
	if err := database.DB.Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
		return nil, errors.New("username already exists")
	}

	// 检查邮箱是否已存在
	if err := database.DB.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		return nil, errors.New("email already exists")
	}

	// 加密密码
	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// 创建用户
	user := &models.User{
		Username: req.Username,
		Email:    req.Email,
		Password: hashedPassword,
		IsActive: true,
		IsAdmin:  false,
	}

	if err := database.DB.Create(user).Error; err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// Login 用户登录
func (s *UserService) Login(req *LoginRequest) (*LoginResponse, error) {
	var user models.User
	if err := database.DB.Where("username = ? AND is_active = ?", req.Username, true).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid username or password")
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	// 验证密码
	if !auth.CheckPassword(req.Password, user.Password) {
		return nil, errors.New("invalid username or password")
	}

	// 生成 JWT Token
	accessToken, err := auth.GenerateToken(user.ID, user.Username, user.IsAdmin)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := auth.GenerateRefreshToken(user.ID, user.Username, user.IsAdmin)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &LoginResponse{
		User:         &user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// RefreshToken 刷新访问令牌
func (s *UserService) RefreshToken(req *RefreshRequest) (*LoginResponse, error) {
	claims, err := auth.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}

	// 验证用户是否仍然有效
	var user models.User
	if err := database.DB.Where("id = ? AND is_active = ?", claims.UserID, true).First(&user).Error; err != nil {
		return nil, errors.New("user not found or inactive")
	}

	// 生成新的访问令牌
	accessToken, err := auth.GenerateToken(user.ID, user.Username, user.IsAdmin)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// 生成新的刷新令牌
	refreshToken, err := auth.GenerateRefreshToken(user.ID, user.Username, user.IsAdmin)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &LoginResponse{
		User:         &user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// GetUserByID 根据 ID 获取用户
func (s *UserService) GetUserByID(userID uuid.UUID) (*models.User, error) {
	var user models.User
	if err := database.DB.Where("id = ? AND is_active = ?", userID, true).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("database error: %w", err)
	}
	return &user, nil
}

// RegenerateAPIKey 重新生成 API 密钥
func (s *UserService) RegenerateAPIKey(userID uuid.UUID) (string, error) {
	var user models.User
	if err := database.DB.Where("id = ? AND is_active = ?", userID, true).First(&user).Error; err != nil {
		return "", errors.New("user not found")
	}

	// 生成新的 API 密钥
	newAPIKey := uuid.New().String() + uuid.New().String()

	if err := database.DB.Model(&user).Update("api_key", newAPIKey).Error; err != nil {
		return "", fmt.Errorf("failed to update API key: %w", err)
	}

	return newAPIKey, nil
}

// UpdateProfile 更新用户资料
func (s *UserService) UpdateProfile(userID uuid.UUID, username, email string) (*models.User, error) {
	var user models.User
	if err := database.DB.Where("id = ? AND is_active = ?", userID, true).First(&user).Error; err != nil {
		return nil, errors.New("user not found")
	}

	// 检查用户名是否被其他用户使用
	if username != user.Username {
		var existingUser models.User
		if err := database.DB.Where("username = ? AND id != ?", username, userID).First(&existingUser).Error; err == nil {
			return nil, errors.New("username already exists")
		}
	}

	// 检查邮箱是否被其他用户使用
	if email != user.Email {
		var existingUser models.User
		if err := database.DB.Where("email = ? AND id != ?", email, userID).First(&existingUser).Error; err == nil {
			return nil, errors.New("email already exists")
		}
	}

	// 更新用户信息
	user.Username = username
	user.Email = email

	if err := database.DB.Save(&user).Error; err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return &user, nil
}

// ChangePassword 修改密码
func (s *UserService) ChangePassword(userID uuid.UUID, oldPassword, newPassword string) error {
	var user models.User
	if err := database.DB.Where("id = ? AND is_active = ?", userID, true).First(&user).Error; err != nil {
		return errors.New("user not found")
	}

	// 验证旧密码
	if !auth.CheckPassword(oldPassword, user.Password) {
		return errors.New("invalid old password")
	}

	// 加密新密码
	hashedPassword, err := auth.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// 更新密码
	if err := database.DB.Model(&user).Update("password", hashedPassword).Error; err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

// UpgradeUserPlan 升级用户计划
func (s *UserService) UpgradeUserPlan(userID string, newPlanType string, expiresAt *time.Time) error {
	// 获取用户当前信息
	var user models.User
	if err := s.db.First(&user, "id = ?", userID).Error; err != nil {
		return fmt.Errorf("用户不存在: %v", err)
	}

	// 获取新计划的保留天数
	retentionDays := s.getPlanRetentionDays(newPlanType)

	// 转换 userID 为 UUID
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("无效的用户ID: %v", err)
	}

	// 开始事务
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 更新用户订阅信息
	subscription := models.UserSubscription{
		UserID:    userUUID,
		PlanType:  models.PlanType(newPlanType),
		Status:    "active",
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 检查是否已有订阅记录
	var existingSubscription models.UserSubscription
	if err := tx.Where("user_id = ?", userID).First(&existingSubscription).Error; err == nil {
		// 更新现有订阅
		if err := tx.Model(&existingSubscription).Updates(map[string]interface{}{
			"plan_type":  newPlanType,
			"status":     "active",
			"expires_at": expiresAt,
			"updated_at": time.Now(),
		}).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("更新订阅失败: %v", err)
		}
	} else {
		// 创建新订阅
		if err := tx.Create(&subscription).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("创建订阅失败: %v", err)
		}
	}

	// 更新用户的现有内容过期时间（升级时延长）
	if err := s.updateUserContentExpiration(tx, userID, retentionDays, true); err != nil {
		tx.Rollback()
		return fmt.Errorf("更新内容过期时间失败: %v", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("提交事务失败: %v", err)
	}

	log.Printf("用户 %s 成功升级到 %s 计划", userID, newPlanType)
	return nil
}

// DowngradeUserPlan 降级用户计划
func (s *UserService) DowngradeUserPlan(userID string, newPlanType string) error {
	// 获取用户当前信息
	var user models.User
	if err := s.db.First(&user, "id = ?", userID).Error; err != nil {
		return fmt.Errorf("用户不存在: %v", err)
	}

	// 获取新计划的保留天数
	retentionDays := s.getPlanRetentionDays(newPlanType)

	// 开始事务
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 更新用户订阅信息
	var subscription models.UserSubscription
	if err := tx.Where("user_id = ?", userID).First(&subscription).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("用户订阅不存在: %v", err)
	}

	// 更新订阅状态
	if err := tx.Model(&subscription).Updates(map[string]interface{}{
		"plan_type":  newPlanType,
		"status":     "active",
		"expires_at": nil, // 降级通常不设置过期时间
		"updated_at": time.Now(),
	}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("更新订阅失败: %v", err)
	}

	// 更新用户的现有内容过期时间（降级时缩短）
	if err := s.updateUserContentExpiration(tx, userID, retentionDays, false); err != nil {
		tx.Rollback()
		return fmt.Errorf("更新内容过期时间失败: %v", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("提交事务失败: %v", err)
	}

	log.Printf("用户 %s 成功降级到 %s 计划", userID, newPlanType)
	return nil
}

// updateUserContentExpiration 更新用户内容的过期时间
func (s *UserService) updateUserContentExpiration(tx *gorm.DB, userID string, retentionDays int, isUpgrade bool) error {
	// 获取用户的所有活跃内容
	var contents []models.Content
	if err := tx.Where("user_id = ? AND is_active = true", userID).Find(&contents).Error; err != nil {
		return fmt.Errorf("获取用户内容失败: %v", err)
	}

	now := time.Now()

	for _, content := range contents {
		var newExpiresAt *time.Time

		if retentionDays > 0 {
			// 计算新的过期时间
			if isUpgrade {
				// 升级：从创建时间重新计算过期时间
				expiresAt := content.CreatedAt.AddDate(0, 0, retentionDays)
				newExpiresAt = &expiresAt
			} else {
				// 降级：从当前时间计算过期时间，但不能早于现在
				expiresAt := now.AddDate(0, 0, retentionDays)
				if content.ExpiresAt != nil && content.ExpiresAt.Before(expiresAt) {
					// 如果原过期时间更早，保持原时间
					newExpiresAt = content.ExpiresAt
				} else {
					newExpiresAt = &expiresAt
				}
			}
		} else {
			// 无限期保存
			newExpiresAt = nil
		}

		// 更新内容过期时间
		if err := tx.Model(&content).Update("expires_at", newExpiresAt).Error; err != nil {
			return fmt.Errorf("更新内容 %s 过期时间失败: %v", content.ID, err)
		}
	}

	return nil
}

// getPlanRetentionDays 获取计划的文章保留天数
func (s *UserService) getPlanRetentionDays(planType string) int {
	switch planType {
	case "community":
		return 7
	case "developer":
		return 30
	case "pro":
		return 90
	case "max":
		return 365
	case "enterprise":
		return -1 // 无限期
	default:
		return 7 // 默认为社区版
	}
}
