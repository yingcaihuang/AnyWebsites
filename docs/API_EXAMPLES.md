# AnyWebsites API 使用示例

## 📋 概述

本文档提供了 AnyWebsites API 的详细使用示例，包括完整的请求/响应示例、错误处理和最佳实践。

## 🚀 快速开始

### 1. 用户注册和登录流程

#### 步骤 1: 注册新用户

```bash
curl -X POST https://localhost:8443/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "developer",
    "email": "developer@example.com",
    "password": "securepassword123"
  }'
```

**成功响应示例：**
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "username": "developer",
  "email": "developer@example.com",
  "api_key": "ak_1234567890abcdef",
  "is_active": true,
  "is_admin": false,
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:30:00Z"
}
```

#### 步骤 2: 用户登录

```bash
curl -X POST https://localhost:8443/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "developer",
    "password": "securepassword123"
  }'
```

**成功响应示例：**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_in": 86400,
  "user": {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "username": "developer",
    "email": "developer@example.com",
    "api_key": "ak_1234567890abcdef",
    "is_active": true,
    "is_admin": false
  }
}
```

### 2. 上传和管理 HTML 内容

#### 上传公开内容

```bash
curl -X POST https://localhost:8443/api/content/upload \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json" \
  -d '{
    "title": "我的个人主页",
    "description": "这是我的个人介绍页面",
    "html_content": "<!DOCTYPE html><html><head><title>Hello World</title></head><body><h1>欢迎访问我的页面！</h1><p>这是一个示例页面。</p></body></html>",
    "is_public": true
  }'
```

**成功响应示例：**
```json
{
  "message": "Content uploaded successfully",
  "content": {
    "id": "456e7890-e89b-12d3-a456-426614174001",
    "user_id": "123e4567-e89b-12d3-a456-426614174000",
    "title": "我的个人主页",
    "description": "这是我的个人介绍页面",
    "access_code": "",
    "is_public": true,
    "expires_at": null,
    "view_count": 0,
    "is_active": true,
    "created_at": "2024-01-15T11:00:00Z",
    "updated_at": "2024-01-15T11:00:00Z"
  }
}
```

#### 上传私有内容（带访问码）

```bash
curl -X POST https://localhost:8443/api/content/upload \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json" \
  -d '{
    "title": "私密文档",
    "description": "仅限内部访问的文档",
    "html_content": "<!DOCTYPE html><html><head><title>Private Doc</title></head><body><h1>机密信息</h1><p>这是私密内容。</p></body></html>",
    "is_public": false,
    "access_code": "secret123",
    "expires_at": "2024-12-31T23:59:59Z"
  }'
```

#### 获取内容列表

```bash
curl -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
     "https://localhost:8443/api/content?page=1&limit=10"
```

**成功响应示例：**
```json
{
  "contents": [
    {
      "id": "456e7890-e89b-12d3-a456-426614174001",
      "title": "我的个人主页",
      "description": "这是我的个人介绍页面",
      "is_public": true,
      "view_count": 15,
      "created_at": "2024-01-15T11:00:00Z"
    }
  ],
  "total": 1,
  "page": 1,
  "limit": 10
}
```

### 3. 访问发布的页面

#### 访问公开页面

```bash
# 直接在浏览器中访问
https://localhost:8443/view/456e7890-e89b-12d3-a456-426614174001

# 或使用 curl
curl https://localhost:8443/view/456e7890-e89b-12d3-a456-426614174001
```

#### 访问私有页面（需要访问码）

```bash
# 在浏览器中访问（会提示输入访问码）
https://localhost:8443/view/789e0123-e89b-12d3-a456-426614174002

# 或直接在 URL 中提供访问码
https://localhost:8443/view/789e0123-e89b-12d3-a456-426614174002?code=secret123

# 使用 curl
curl "https://localhost:8443/view/789e0123-e89b-12d3-a456-426614174002?code=secret123"
```

### 4. 使用 API Key 认证

除了 JWT Token，您也可以使用 API Key 进行认证：

#### 通过请求头

```bash
curl -H "X-API-Key: ak_1234567890abcdef" \
     https://localhost:8443/api/content
```

#### 通过查询参数

```bash
curl "https://localhost:8443/api/content?api_key=ak_1234567890abcdef"
```

## 🔧 管理后台 API 示例

### 获取地理位置统计

```bash
curl -b "admin_session=<session-id>" \
     "https://localhost:8443/admin/api/geoip-stats?range=7d"
```

**响应示例：**
```json
{
  "success": true,
  "stats": [
    {
      "country": "中国",
      "region": "北京市",
      "city": "北京",
      "count": 25
    },
    {
      "country": "美国",
      "region": "加利福尼亚州",
      "city": "旧金山",
      "count": 18
    }
  ]
}
```

### 获取系统设置

```bash
curl -b "admin_session=<session-id>" \
     https://localhost:8443/admin/api/settings
```

## ❌ 错误处理

### 常见错误响应

#### 401 未授权

```json
{
  "success": false,
  "error": "Unauthorized: Invalid or expired token"
}
```

#### 400 请求参数错误

```json
{
  "success": false,
  "error": "Validation failed: username is required"
}
```

#### 404 资源不存在

```json
{
  "success": false,
  "error": "Content not found"
}
```

#### 409 资源冲突

```json
{
  "success": false,
  "error": "Username already exists"
}
```

## 💡 最佳实践

### 1. 令牌管理

- 访问令牌有效期为 24 小时，请及时刷新
- 刷新令牌有效期为 7 天
- 建议在令牌过期前 1 小时进行刷新

### 2. 错误重试

- 对于 5xx 错误，建议使用指数退避重试
- 对于 429 限流错误，请遵守 Retry-After 头部
- 对于 4xx 错误，请检查请求参数后重试

### 3. 安全建议

- 始终使用 HTTPS 进行 API 调用
- 不要在客户端代码中硬编码 API 密钥
- 定期轮换 API 密钥
- 使用最小权限原则

### 4. 性能优化

- 使用分页参数控制返回数据量
- 合理设置请求超时时间
- 对于频繁访问的数据，考虑客户端缓存

## 📞 技术支持

如有问题或需要技术支持，请联系：
- 📧 邮箱: support@anywebsites.com
- 📖 文档: [Swagger UI](https://localhost:8443/docs/swagger-ui.html)
- 🔧 工具: [Postman 集合](https://localhost:8443/docs/AnyWebsites-API.postman_collection.json)

---

*最后更新: 2025-06-24*
