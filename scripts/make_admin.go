package main

import (
	"fmt"
	"log"
	"os"

	"anywebsites/internal/config"
	"anywebsites/internal/database"
	"anywebsites/internal/models"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run scripts/make_admin.go <username>")
		os.Exit(1)
	}

	username := os.Args[1]

	// 加载配置
	cfg := config.Load()

	// 连接数据库
	if err := database.Connect(cfg); err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// 查找用户
	var user models.User
	if err := database.DB.Where("username = ?", username).First(&user).Error; err != nil {
		log.Fatal("User not found:", err)
	}

	// 设置为管理员
	user.IsAdmin = true
	if err := database.DB.Save(&user).Error; err != nil {
		log.Fatal("Failed to update user:", err)
	}

	fmt.Printf("User '%s' is now an admin!\n", username)
}
