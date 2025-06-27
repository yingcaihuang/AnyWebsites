# AnyWebsites API 文档

## 📖 概述

AnyWebsites 是一个功能强大的 HTML 页面托管服务平台，基于 Golang 开发，提供完整的 RESTful API 接口。本平台专为开发者和内容创作者设计，支持快速部署、管理和分析 HTML 页面。

### 🌟 核心特性

- **🚀 快速部署**: 一键上传 HTML 内容，自动生成访问链接
- **🔐 安全可靠**: 多重认证机制，数据加密存储
- **📊 智能分析**: 实时访问统计，地理位置分析
- **⚙️ 灵活配置**: 动态系统设置，热重载支持
- **👥 用户管理**: 完整的用户权限体系
- **🌍 全球部署**: 支持 Docker 容器化部署

### 🎯 适用场景

- **个人作品展示**: 快速发布个人项目和作品
- **临时页面托管**: 活动页面、落地页等临时需求
- **原型演示**: 前端原型快速展示和分享
- **静态网站**: 简单的静态网站托管
- **API 文档**: 在线文档和说明页面

## 🚀 快速开始

### 访问 API 文档

- **Swagger UI**: [https://localhost/docs/swagger-ui.html](https://localhost/docs/swagger-ui.html)
- **OpenAPI 规范**: [https://localhost/docs/swagger.yaml](https://localhost/docs/swagger.yaml)

### 基础信息

- **API 基础 URL**: `https://localhost`
- **API 版本**: v1.0.0
- **支持格式**: JSON
- **字符编码**: UTF-8

## 🔐 认证方式

### 1. Bearer Token (JWT)
用于用户 API 访问，通过登录接口获取。

```bash
curl -H "Authorization: Bearer <your-jwt-token>" \
     https://localhost/api/content
```

### 2. API Key
通过请求头或查询参数传递。

```bash
# 请求头方式
curl -H "X-API-Key: <your-api-key>" \
     https://localhost/api/content

# 查询参数方式
curl "https://localhost/api/content?api_key=<your-api-key>"
```

### 3. Admin Session
管理后台会话认证，通过 Cookie 传递。

```bash
curl -b "admin_session=<session-id>" \
     https://localhost/admin/api/users
```

## 📋 API 分类

### 🔐 认证接口 (Authentication)
- `POST /api/auth/register` - 用户注册
- `POST /api/auth/login` - 用户登录
- `POST /api/auth/refresh` - 刷新令牌

### 📄 内容管理 (Content Management)
- `POST /api/content/upload` - 上传 HTML 内容
- `GET /api/content` - 获取内容列表
- `GET /api/content/{id}` - 获取内容详情
- `PUT /api/content/{id}` - 更新内容
- `DELETE /api/content/{id}` - 删除内容

### 🌐 内容访问 (Content Access)
- `GET /view/{id}` - 访问发布的 HTML 页面

### 👥 管理后台 - 用户管理 (Admin - User Management)
- `POST /admin/api/users/{id}/toggle-status` - 切换用户状态
- `POST /admin/api/users/{id}/toggle-admin` - 切换管理员权限
- `POST /admin/api/users/{id}/reset-api-key` - 重置 API 密钥
- `GET /admin/api/users/{id}/details` - 获取用户详情
- `POST /admin/api/users/{id}/reset-password` - 重置密码
- `DELETE /admin/api/users/{id}` - 删除用户

### 📊 管理后台 - 内容管理 (Admin - Content Management)
- `DELETE /admin/api/contents/{id}` - 删除内容
- `POST /admin/api/contents/{id}/restore` - 恢复内容

### 📈 管理后台 - 统计分析 (Admin - Analytics)
- `GET /admin/api/geoip-stats` - 获取地理位置统计

### ⚙️ 管理后台 - 系统设置 (Admin - Settings)
- `GET /admin/api/settings` - 获取所有设置
- `POST /admin/api/settings` - 创建设置
- `GET /admin/api/settings/categories` - 获取设置分类
- `GET /admin/api/settings/category/{category}` - 按分类获取设置
- `PUT /admin/api/settings/{id}` - 更新设置
- `DELETE /admin/api/settings/{id}` - 删除设置
- `GET /admin/api/settings/{category}/{key}/history` - 获取设置历史
- `GET /admin/api/settings/export` - 导出设置
- `POST /admin/api/settings/import` - 导入设置
- `POST /admin/api/settings/reload` - 重载配置

### 🏥 健康检查 (Health)
- `GET /health` - 服务器健康检查

## 💡 使用示例

### 用户注册和登录

```bash
# 1. 注册用户
curl -X POST https://localhost/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@example.com",
    "password": "password123"
  }'

# 2. 用户登录
curl -X POST https://localhost/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "password123"
  }'
```

### 上传和管理内容

```bash
# 1. 上传 HTML 内容
curl -X POST https://localhost/api/content/upload \
  -H "Authorization: Bearer <your-token>" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "我的网页",
    "description": "这是一个测试页面",
    "html_content": "<html><body><h1>Hello World!</h1></body></html>",
    "is_public": true
  }'

# 2. 获取内容列表
curl -H "Authorization: Bearer <your-token>" \
     https://localhost/api/content

# 3. 访问发布的页面
curl https://localhost/view/<content-id>
```

### 管理后台操作

```bash
# 1. 获取用户详情（需要管理员权限）
curl -b "admin_session=<session-id>" \
     https://localhost/admin/api/users/<user-id>/details

# 2. 获取系统设置
curl -b "admin_session=<session-id>" \
     https://localhost/admin/api/settings

# 3. 获取地理位置统计
curl -b "admin_session=<session-id>" \
     "https://localhost/admin/api/geoip-stats?range=7d"
```

## 📝 响应格式

### 成功响应
```json
{
  "success": true,
  "data": { ... },
  "message": "操作成功"
}
```

### 错误响应
```json
{
  "success": false,
  "error": "错误信息",
  "code": "ERROR_CODE"
}
```

## 🔧 状态码

- `200` - 请求成功
- `201` - 创建成功
- `400` - 请求参数错误
- `401` - 未授权访问
- `403` - 权限不足
- `404` - 资源不存在
- `409` - 资源冲突
- `500` - 服务器内部错误

## 📚 数据模型

详细的数据模型定义请参考 [Swagger UI](https://localhost/docs/swagger-ui.html) 中的 "Schemas" 部分。

主要模型包括：
- **User** - 用户信息
- **Content** - 内容信息
- **ContentAnalytics** - 访问统计
- **SettingResponse** - 系统设置
- **GeoStats** - 地理位置统计

## 🛠️ 开发工具

### Postman 集合
可以将 OpenAPI 规范导入到 Postman 中：
1. 打开 Postman
2. 点击 "Import"
3. 输入 URL: `https://localhost/docs/swagger.yaml`

### cURL 脚本
所有 API 端点都可以通过 cURL 进行测试，具体示例请参考 Swagger UI 中的 "Try it out" 功能。

## 📞 支持

如有问题或建议，请联系：
- 邮箱: support@anywebsites.com
- 文档: [Swagger UI](https://localhost/docs/swagger-ui.html)

---

*最后更新: 2025-06-24*
