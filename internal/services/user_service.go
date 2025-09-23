package services

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"gpt-load/internal/models"
	"gpt-load/internal/utils"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserService struct {
	db *gorm.DB
}

func NewUserService(db *gorm.DB) *UserService {
	return &UserService{db: db}
}

// CreateUser 创建新用户
func (s *UserService) CreateUser(req CreateUserRequest) (*models.User, error) {
	// 检查用户名和邮箱是否已存在
	var count int64
	s.db.Model(&models.User{}).Where("username = ? OR email = ?", req.Username, req.Email).Count(&count)
	if count > 0 {
		return nil, errors.New("用户名或邮箱已存在")
	}

	// 密码哈希
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("密码加密失败: %w", err)
	}

	user := &models.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		DisplayName:  req.DisplayName,
		Role:         req.Role,
		Status:       models.UserStatusActive,
	}

	if err := s.db.Create(user).Error; err != nil {
		return nil, fmt.Errorf("创建用户失败: %w", err)
	}

	// 记录活动日志
	s.logActivity(user.ID, "user.created", "user", fmt.Sprint(user.ID), "", "", nil)

	return user, nil
}

// Authenticate 用户认证
func (s *UserService) Authenticate(username, password, ipAddress, userAgent string) (*models.User, string, error) {
	// 查找用户
	var user models.User
	err := s.db.Where("username = ? OR email = ?", username, username).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "", errors.New("用户名或密码错误")
		}
		return nil, "", fmt.Errorf("查询用户失败: %w", err)
	}

	// 检查用户状态
	if user.Status != models.UserStatusActive {
		return nil, "", errors.New("用户账户已被禁用")
	}

	// 检查是否被锁定
	if user.LockedUntil != nil && user.LockedUntil.After(time.Now()) {
		return nil, "", fmt.Errorf("账户已被锁定，解锁时间: %s", user.LockedUntil.Format("2006-01-02 15:04:05"))
	}

	// 验证密码
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		// 密码错误，增加失败尝试次数
		s.incrementFailedAttempts(&user)
		return nil, "", errors.New("用户名或密码错误")
	}

	// 重置失败尝试次数
	s.resetFailedAttempts(&user)

	// 更新登录信息
	now := time.Now()
	s.db.Model(&user).Updates(map[string]interface{}{
		"last_login_at": now,
		"last_login_ip": ipAddress,
		"login_count":   gorm.Expr("login_count + 1"),
	})

	// 生成会话令牌
	token, err := s.createSession(&user, ipAddress, userAgent)
	if err != nil {
		return nil, "", fmt.Errorf("创建会话失败: %w", err)
	}

	// 记录活动日志
	s.logActivity(user.ID, "user.login", "", "", ipAddress, userAgent, nil)

	return &user, token, nil
}

// ValidateToken 验证令牌
func (s *UserService) ValidateToken(token string) (*models.User, error) {
	var session models.UserSession
	err := s.db.Preload("User").Where("token = ? AND expires_at > ?", token, time.Now()).First(&session).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("无效的令牌")
		}
		return nil, fmt.Errorf("验证令牌失败: %w", err)
	}

	// 检查用户状态
	if session.User.Status != models.UserStatusActive {
		return nil, errors.New("用户账户已被禁用")
	}

	// 更新会话最后使用时间
	s.db.Model(&session).Update("updated_at", time.Now())

	return &session.User, nil
}

// ListUsers 获取用户列表
func (s *UserService) ListUsers(page, pageSize int, role, status, search string) ([]*models.User, int64, error) {
	query := s.db.Model(&models.User{})

	// 过滤条件
	if role != "" {
		query = query.Where("role = ?", role)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if search != "" {
		query = query.Where("username LIKE ? OR email LIKE ? OR display_name LIKE ?",
			"%"+search+"%", "%"+search+"%", "%"+search+"%")
	}

	// 获取总数
	var total int64
	query.Count(&total)

	// 分页查询
	var users []*models.User
	err := query.Offset((page - 1) * pageSize).Limit(pageSize).Order("created_at DESC").Find(&users).Error
	if err != nil {
		return nil, 0, fmt.Errorf("查询用户列表失败: %w", err)
	}

	return users, total, nil
}

// UpdateUser 更新用户信息
func (s *UserService) UpdateUser(userID uint, req UpdateUserRequest) error {
	updates := make(map[string]interface{})

	if req.DisplayName != nil {
		updates["display_name"] = *req.DisplayName
	}
	if req.Email != nil {
		// 检查邮箱是否已被其他用户使用
		var count int64
		s.db.Model(&models.User{}).Where("email = ? AND id != ?", *req.Email, userID).Count(&count)
		if count > 0 {
			return errors.New("邮箱已被其他用户使用")
		}
		updates["email"] = *req.Email
	}
	if req.Role != nil {
		updates["role"] = *req.Role
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.Avatar != nil {
		updates["avatar"] = *req.Avatar
	}

	if len(updates) == 0 {
		return errors.New("没有要更新的字段")
	}

	err := s.db.Model(&models.User{}).Where("id = ?", userID).Updates(updates).Error
	if err != nil {
		return fmt.Errorf("更新用户失败: %w", err)
	}

	// 记录活动日志
	s.logActivity(userID, "user.updated", "user", fmt.Sprint(userID), "", "", updates)

	return nil
}

// ChangePassword 修改密码
func (s *UserService) ChangePassword(userID uint, oldPassword, newPassword string) error {
	var user models.User
	err := s.db.First(&user, userID).Error
	if err != nil {
		return fmt.Errorf("用户不存在: %w", err)
	}

	// 验证旧密码
	if oldPassword != "" {
		err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword))
		if err != nil {
			return errors.New("原密码错误")
		}
	}

	// 生成新密码哈希
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("密码加密失败: %w", err)
	}

	// 更新密码
	err = s.db.Model(&user).Update("password_hash", string(hashedPassword)).Error
	if err != nil {
		return fmt.Errorf("更新密码失败: %w", err)
	}

	// 记录活动日志
	s.logActivity(userID, "user.password_changed", "user", fmt.Sprint(userID), "", "", nil)

	return nil
}

// DeleteUser 删除用户
func (s *UserService) DeleteUser(userID uint) error {
	// 删除用户会话
	s.db.Where("user_id = ?", userID).Delete(&models.UserSession{})

	// 删除用户分组关联
	s.db.Where("user_id = ?", userID).Delete(&models.UserGroup{})

	// 删除用户
	err := s.db.Delete(&models.User{}, userID).Error
	if err != nil {
		return fmt.Errorf("删除用户失败: %w", err)
	}

	// 记录活动日志
	s.logActivity(userID, "user.deleted", "user", fmt.Sprint(userID), "", "", nil)

	return nil
}

// Logout 用户退出
func (s *UserService) Logout(token string) error {
	return s.db.Where("token = ?", token).Delete(&models.UserSession{}).Error
}

// 内部方法

func (s *UserService) incrementFailedAttempts(user *models.User) {
	user.FailedAttempts++

	// 如果连续失败5次，锁定账户30分钟
	if user.FailedAttempts >= 5 {
		lockUntil := time.Now().Add(30 * time.Minute)
		user.LockedUntil = &lockUntil
	}

	s.db.Model(user).Updates(map[string]interface{}{
		"failed_attempts": user.FailedAttempts,
		"locked_until":    user.LockedUntil,
	})
}

func (s *UserService) resetFailedAttempts(user *models.User) {
	s.db.Model(user).Updates(map[string]interface{}{
		"failed_attempts": 0,
		"locked_until":    nil,
	})
}

func (s *UserService) createSession(user *models.User, ipAddress, userAgent string) (string, error) {
	// 生成会话ID和令牌
	sessionID := utils.GenerateID()
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", err
	}
	token := hex.EncodeToString(tokenBytes)

	// 创建会话记录
	session := &models.UserSession{
		ID:        sessionID,
		UserID:    user.ID,
		Token:     token,
		UserAgent: userAgent,
		IPAddress: ipAddress,
		ExpiresAt: time.Now().Add(24 * time.Hour), // 24小时过期
	}

	err := s.db.Create(session).Error
	if err != nil {
		return "", err
	}

	return token, nil
}

func (s *UserService) logActivity(userID uint, action, resource, resourceID, ipAddress, userAgent string, details map[string]interface{}) {
	activity := &models.UserActivity{
		UserID:     userID,
		Action:     action,
		Resource:   resource,
		ResourceID: resourceID,
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
		Timestamp:  time.Now(),
	}

	if details != nil {
		activity.Details = utils.ToJSON(details)
	}

	s.db.Create(activity)
}

// 请求和响应结构

type CreateUserRequest struct {
	Username    string `json:"username" binding:"required,min=3,max=50"`
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required,min=6"`
	DisplayName string `json:"display_name"`
	Role        string `json:"role" binding:"required,oneof=admin user viewer"`
}

type UpdateUserRequest struct {
	DisplayName *string `json:"display_name"`
	Email       *string `json:"email"`
	Role        *string `json:"role"`
	Status      *string `json:"status"`
	Avatar      *string `json:"avatar"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}