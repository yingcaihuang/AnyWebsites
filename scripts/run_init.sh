#!/bin/bash

# 用户计划初始化脚本

echo "🚀 开始初始化用户计划系统..."

# 设置数据库连接
export DATABASE_URL="host=localhost user=anywebsites password=anywebsites dbname=anywebsites port=5432 sslmode=disable TimeZone=UTC"

# 进入脚本目录
cd "$(dirname "$0")"

# 运行初始化脚本
echo "📦 运行初始化脚本..."
go run init_user_plans.go

echo "🎉 用户计划系统初始化完成！"
