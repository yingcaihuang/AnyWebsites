# AnyWebsites Docker 镜像清单

## 镜像仓库信息

**仓库地址**: hub.verycloud.cn  
**项目名称**: anywebsites  
**推送时间**: 2025-07-11

## 镜像列表

### 1. 应用程序镜像 (App)

**镜像名称**: `hub.verycloud.cn/anywebsites/app`

| 标签 | 镜像ID | 大小 | 描述 |
|------|--------|------|------|
| 1.2.0 | 9daa95a52d8a | 37.3MB | 版本1.2.0，包含最新功能和修复 |
| latest | 9daa95a52d8a | 37.3MB | 最新版本（指向1.2.0） |

**功能特性**:
- ✅ 内容管理系统
- ✅ 用户管理和认证
- ✅ API密钥管理
- ✅ 地理位置分析（GeoIP）
- ✅ 访问统计和分析
- ✅ 计划管理系统
- ✅ Redis缓存支持
- ✅ PostgreSQL数据库支持
- ✅ JWT认证
- ✅ 健康检查端点

### 2. Nginx镜像

**镜像名称**: `hub.verycloud.cn/anywebsites/nginx`

| 标签 | 镜像ID | 大小 | 描述 |
|------|--------|------|------|
| 1.2.0 | 54bcf97d500f | 53.9MB | 版本1.2.0，修复HTTP/2配置警告 |
| latest | 54bcf97d500f | 53.9MB | 最新版本（指向1.2.0） |

**配置特性**:
- ✅ HTTP/2 支持（新语法）
- ✅ SSL/TLS 加密
- ✅ 自动SSL证书生成
- ✅ 反向代理配置
- ✅ 静态文件服务
- ✅ 安全头配置
- ✅ 访问日志记录
- ✅ 支持多域名配置

## 版本更新记录

### v1.2.0 (2025-07-11)

**应用程序更新**:
- 🔧 修复内容添加功能中的字段名不匹配问题
- 🔧 统一使用AccessCount替代ViewCount
- 🔧 修复表单字段名称一致性问题
- ✨ 完善内容管理功能
- ✨ 改进批量操作功能

**Nginx更新**:
- 🔧 修复HTTP/2配置警告（使用新语法）
- 🔧 将`listen 443 ssl http2;`更新为`listen 443 ssl; http2 on;`
- ✨ 保持HTTP/2功能正常工作
- ✨ 消除nginx启动时的弃用警告

## 使用方法

### 拉取镜像

```bash
# 拉取应用程序镜像
docker pull hub.verycloud.cn/anywebsites/app:1.2.0
docker pull hub.verycloud.cn/anywebsites/app:latest

# 拉取Nginx镜像
docker pull hub.verycloud.cn/anywebsites/nginx:1.2.0
docker pull hub.verycloud.cn/anywebsites/nginx:latest
```

### Docker Compose 使用

```yaml
version: '3.8'
services:
  app:
    image: hub.verycloud.cn/anywebsites/app:1.2.0
    # ... 其他配置
  
  nginx:
    image: hub.verycloud.cn/anywebsites/nginx:1.2.0
    # ... 其他配置
```

## 镜像构建信息

### 应用程序镜像构建

- **基础镜像**: Alpine Linux
- **Go版本**: 最新稳定版
- **构建方式**: 多阶段构建
- **优化**: 最小化镜像大小

### Nginx镜像构建

- **基础镜像**: nginx:alpine
- **SSL支持**: OpenSSL
- **配置文件**: 自定义nginx.conf
- **证书生成**: 自动生成自签名证书

## 安全说明

### 默认账户

**管理员账户**:
- 用户名: `admin`, 密码: `Google@google`
- 用户名: `yingcai`, 密码: `Yingcai@yingcai`

⚠️ **生产环境部署时请务必修改默认密码！**

### 环境变量

生产环境请修改以下敏感配置：
- `JWT_SECRET`: JWT签名密钥
- `DB_PASSWORD`: 数据库密码
- `ADMIN_PASSWORD`: 管理员密码

## 技术支持

如有问题，请联系：
- 项目维护者
- 技术支持团队

## 许可证

请参考项目根目录的LICENSE文件。
