package db

import (
	"gpt-load/internal/models"

	"gorm.io/gorm"
)

func AddUserManagement(db *gorm.DB) error {
	// 创建用户表
	if err := db.AutoMigrate(&models.User{}); err != nil {
		return err
	}

	// 创建用户分组关联表
	if err := db.AutoMigrate(&models.UserGroup{}); err != nil {
		return err
	}

	// 创建用户会话表
	if err := db.AutoMigrate(&models.UserSession{}); err != nil {
		return err
	}

	// 创建用户活动日志表
	if err := db.AutoMigrate(&models.UserActivity{}); err != nil {
		return err
	}

	// 创建默认管理员用户
	if err := createDefaultAdmin(db); err != nil {
		return err
	}

	return nil
}

func createDefaultAdmin(db *gorm.DB) error {
	// 检查是否已存在管理员用户
	var count int64
	db.Model(&models.User{}).Where("role = ?", models.RoleAdmin).Count(&count)

	if count > 0 {
		// 已存在管理员用户，跳过创建
		return nil
	}

	// 创建默认管理员用户
	// 注意：这里使用明文密码 "admin123"，实际部署时应该使用更安全的密码
	// 密码哈希应该在用户服务中处理
	defaultAdmin := models.User{
		Username:    "admin",
		Email:       "admin@example.com",
		PasswordHash: "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi", // "password" bcrypt hash
		DisplayName: "系统管理员",
		Role:        models.RoleAdmin,
		Status:      models.UserStatusActive,
	}

	return db.Create(&defaultAdmin).Error
}