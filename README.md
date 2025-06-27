# AnyWebsites - HTML 页面托管服务

一个基于 Golang 的 HTML 页面托管服务，支持用户上传、管理和统计功能。

## 功能特性

- 🚀 **HTML 页面上传**: 通过 API 或后台页面上传 HTML 代码
- 🔐 **用户管理**: 完整的用户注册、登录和密钥管理系统
- ⏰ **过期控制**: 支持设置内容过期时间和加密访问
- 📊 **统计分析**: 访问来源、流量、请求数、地理位置统计
- 🛡️ **API 鉴权**: Bearer Token 认证方式
- 🐳 **容器化部署**: 支持 Docker Compose 一键部署
- 🌐 **Nginx 代理**: 高性能的静态文件服务

## 技术栈

- **后端**: Go 1.21 + Gin Framework
- **数据库**: PostgreSQL + Redis
- **认证**: JWT Token
- **部署**: Docker + Docker Compose + Nginx
- **地理位置**: GeoIP2

## 快速开始

### 1. 克隆项目

```bash
git clone <repository-url>
cd AnyWebsites
```

### 2. 环境配置

```bash
cp .env.example .env
# 编辑 .env 文件，配置数据库和其他参数
```

### 3. Docker 部署

```bash
docker-compose up -d
```

### 4. 本地开发

```bash
# 安装依赖
go mod tidy

# 运行服务
go run cmd/server/main.go
```

## API 文档

### 认证相关

- `POST /api/auth/register` - 用户注册
- `POST /api/auth/login` - 用户登录
- `POST /api/auth/refresh` - 刷新 Token

### 内容管理

- `POST /api/content/upload` - 上传 HTML 内容
- `GET /api/content` - 获取内容列表
- `GET /api/content/:id` - 获取内容详情
- `PUT /api/content/:id` - 更新内容
- `DELETE /api/content/:id` - 删除内容

### 统计分析

- `GET /api/stats/overview` - 总览统计
- `GET /api/stats/traffic` - 流量统计
- `GET /api/stats/geo` - 地理位置统计

### 内容访问

- `GET /view/:id` - 访问发布的 HTML 页面
- `GET /view/:id/:code` - 加密访问

## 项目结构

```
AnyWebsites/
├── cmd/
│   └── server/          # 主程序入口
├── internal/
│   ├── api/            # API 路由和处理器
│   ├── auth/           # 认证相关
│   ├── config/         # 配置管理
│   ├── database/       # 数据库连接和迁移
│   ├── middleware/     # 中间件
│   ├── models/         # 数据模型
│   ├── services/       # 业务逻辑
│   └── utils/          # 工具函数
├── web/
│   ├── static/         # 静态文件
│   └── templates/      # HTML 模板
├── uploads/            # 上传文件存储
├── docker-compose.yml  # Docker 编排文件
├── Dockerfile         # Docker 镜像构建
└── nginx.conf         # Nginx 配置
```

## 许可证

MIT License
