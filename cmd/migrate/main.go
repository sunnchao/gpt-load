package main

import (
	"gpt-load/internal/container"
	"gpt-load/internal/db/migrations"
	"log"

	"gorm.io/gorm"
)

func main() {
	// 构建依赖注入容器
	c, err := container.BuildContainer()
	if err != nil {
		log.Fatalf("Failed to build container: %v", err)
	}

	// 获取数据库连接
	var db *gorm.DB
	if err := c.Invoke(func(database *gorm.DB) {
		db = database
	}); err != nil {
		log.Fatalf("Failed to get database: %v", err)
	}

	log.Println("开始执行用户管理系统数据库迁移...")

	// 执行用户管理系统迁移
	if err := migrations.AddUserManagement(db); err != nil {
		log.Fatalf("用户管理系统迁移失败: %v", err)
	}

	log.Println("用户管理系统数据库迁移完成!")
	log.Println("默认管理员账户: 用户名=admin, 密码=password")
}