# Nginx 代理集成说明

## 概述

AnyWebsites 现在已经集成了 Nginx 作为前端代理服务器，提供以下功能：

- **HTTPS/SSL 支持**：自动生成自签名证书用于开发测试
- **HTTP 到 HTTPS 重定向**：自动将 HTTP 请求重定向到 HTTPS
- **真实 IP 获取**：正确传递客户端真实 IP 地址给后端应用
- **负载均衡**：为后端应用提供代理和负载均衡
- **静态文件缓存**：优化静态资源的缓存策略

## 架构

```
客户端 → Nginx (端口 8080/8443) → Go 应用 (端口 8085) → PostgreSQL/Redis
```

## 端口配置

- **HTTP**: `localhost:8080` (重定向到 HTTPS)
- **HTTPS**: `localhost:8443` (主要访问端口)
- **后端应用**: `8085` (仅内部网络访问)

## 真实 IP 获取

Nginx 配置了以下头部来传递真实客户端 IP：

1. `X-Real-IP`: 直接的客户端 IP
2. `X-Forwarded-For`: 代理链中的 IP 列表
3. `X-Forwarded-Proto`: 原始协议 (http/https)
4. `X-Forwarded-Host`: 原始主机名

后端应用通过 `getRealClientIP()` 函数按优先级获取真实 IP：
1. X-Real-IP 头部
2. X-Forwarded-For 头部的第一个 IP
3. CF-Connecting-IP 头部 (Cloudflare 支持)
4. Gin 默认的 ClientIP()

## SSL 证书

开发环境使用自签名证书，生产环境建议：
- 使用 Let's Encrypt 免费证书
- 或购买商业 SSL 证书
- 配置证书自动续期

## 启动服务

```bash
# 构建并启动所有服务
docker-compose up --build -d

# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs nginx
docker-compose logs app
```

## 访问地址

- **管理后台**: https://localhost:8443/admin/login
- **API 接口**: https://localhost:8443/api/
- **内容查看**: https://localhost:8443/view/{content-id}
- **健康检查**: https://localhost:8443/health

## 生产环境配置建议

### 1. SSL 证书配置

```nginx
# 使用真实证书替换自签名证书
ssl_certificate /path/to/your/certificate.crt;
ssl_certificate_key /path/to/your/private.key;
```

### 2. 安全头部

```nginx
# 已配置的安全头部
add_header X-Frame-Options DENY;
add_header X-Content-Type-Options nosniff;
add_header X-XSS-Protection "1; mode=block";
add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
```

### 3. 日志配置

```nginx
# 自定义日志格式，包含真实 IP
log_format main '$remote_addr - $remote_user [$time_local] "$request" '
                '$status $body_bytes_sent "$http_referer" '
                '"$http_user_agent" "$http_x_forwarded_for"';
```

### 4. 性能优化

- 启用 Gzip 压缩 ✅
- 配置静态文件缓存 ✅
- 设置适当的 worker 进程数
- 配置连接池和超时

## 监控和日志

### 查看 Nginx 日志
```bash
docker-compose logs nginx
```

### 查看应用日志
```bash
docker-compose logs app
```

### 实时监控
```bash
docker-compose logs -f nginx app
```

## 故障排除

### 1. SSL 证书问题
如果遇到 SSL 证书错误，重新生成证书：
```bash
docker-compose exec nginx /usr/local/bin/generate-ssl.sh
docker-compose restart nginx
```

### 2. 代理连接问题
检查后端应用是否正常运行：
```bash
docker-compose exec app curl http://localhost:8085/health
```

### 3. 真实 IP 获取问题
检查应用日志中的 IP 地址记录，确认是否正确获取真实 IP。

## 扩展功能

### 1. 多域名支持
可以配置多个 server 块支持不同域名。

### 2. API 限流
可以添加 `limit_req` 模块进行 API 限流。

### 3. 缓存策略
可以配置更复杂的缓存策略，包括 API 响应缓存。

### 4. 负载均衡
可以配置多个后端实例进行负载均衡。

## 安全注意事项

1. **生产环境必须使用真实 SSL 证书**
2. **配置防火墙规则，只开放必要端口**
3. **定期更新 Nginx 和相关组件**
4. **监控访问日志，及时发现异常访问**
5. **配置适当的速率限制和 DDoS 防护**

## 总结

通过 Nginx 代理集成，AnyWebsites 现在具备了：
- ✅ 生产级别的 HTTPS 支持
- ✅ 真实 IP 地址获取和地理位置分析
- ✅ 静态文件优化和缓存
- ✅ 安全头部配置
- ✅ 容器化部署支持

这为应用的生产部署奠定了坚实的基础。
