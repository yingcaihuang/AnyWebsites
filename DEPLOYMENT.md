# AnyWebsites 部署指南

## 镜像信息

项目已经构建并推送到 hub.verycloud.cn 的 anywebsites 项目下：

### 应用程序镜像
- `hub.verycloud.cn/anywebsites/app:1.2.0` - 应用程序镜像（版本1.2.0）
- `hub.verycloud.cn/anywebsites/app:latest` - 应用程序镜像（最新版本）

### Nginx镜像
- `hub.verycloud.cn/anywebsites/nginx:1.2.0` - Nginx镜像（版本1.2.0）
- `hub.verycloud.cn/anywebsites/nginx:latest` - Nginx镜像（最新版本）

## 快速部署

### 1. 准备环境

确保目标服务器已安装：
- Docker
- Docker Compose

### 2. 下载部署文件

```bash
# 下载生产环境的docker-compose文件
wget https://raw.githubusercontent.com/your-repo/anywebsites/main/docker-compose.prod.yml

# 下载初始化SQL文件
wget https://raw.githubusercontent.com/your-repo/anywebsites/main/init.sql

# 下载环境变量模板
wget https://raw.githubusercontent.com/your-repo/anywebsites/main/.env.example -O .env
```

### 3. 配置环境变量

编辑 `.env` 文件，设置必要的环境变量：

```bash
# 数据库配置
DB_HOST=postgres
DB_PORT=5432
DB_USER=anywebsites
DB_PASSWORD=anywebsites123  # 生产环境请修改为强密码
DB_NAME=anywebsites

# Redis配置
REDIS_HOST=redis
REDIS_PORT=6379

# JWT密钥（生产环境必须修改）
JWT_SECRET=your-super-secret-jwt-key-change-this-in-production

# 服务端口
ADMIN_PORT=8080
API_PORT=8085

# GeoIP数据库路径
GEOIP_DB_PATH=/app/GeoLite2-City.mmdb
```

### 4. 启动服务

```bash
# 使用生产环境配置启动
docker-compose -f docker-compose.prod.yml up -d

# 查看服务状态
docker-compose -f docker-compose.prod.yml ps

# 查看日志
docker-compose -f docker-compose.prod.yml logs -f
```

### 5. 验证部署

访问以下地址验证部署是否成功：

- HTTP: `http://your-domain.com`
- HTTPS: `https://your-domain.com`
- 管理后台: `https://your-domain.com/admin`

默认管理员账户：
- 用户名: `admin`，密码: `Google@google`
- 用户名: `yingcai`，密码: `Yingcai@yingcai`

## 生产环境注意事项

### 1. 安全配置

- 修改默认的数据库密码
- 修改JWT密钥
- 修改默认管理员密码
- 配置防火墙规则

### 2. SSL证书

生产环境建议使用真实的SSL证书：

```bash
# 将证书文件放置到nginx容器的SSL目录
docker cp your-cert.pem container_name:/etc/nginx/ssl/cert.pem
docker cp your-key.pem container_name:/etc/nginx/ssl/key.pem

# 重启nginx
docker-compose -f docker-compose.prod.yml restart nginx
```

### 3. 数据备份

定期备份PostgreSQL数据：

```bash
# 备份数据库
docker exec anywebsites-postgres-1 pg_dump -U anywebsites anywebsites > backup_$(date +%Y%m%d_%H%M%S).sql

# 恢复数据库
docker exec -i anywebsites-postgres-1 psql -U anywebsites anywebsites < backup.sql
```

### 4. 监控和日志

```bash
# 查看容器状态
docker-compose -f docker-compose.prod.yml ps

# 查看资源使用情况
docker stats

# 查看特定服务日志
docker-compose -f docker-compose.prod.yml logs app
docker-compose -f docker-compose.prod.yml logs nginx
```

## 更新部署

### 1. 拉取最新镜像

```bash
docker pull hub.verycloud.cn/anywebsites/app:latest
docker pull hub.verycloud.cn/anywebsites/nginx:latest
```

### 2. 重新部署

```bash
docker-compose -f docker-compose.prod.yml down
docker-compose -f docker-compose.prod.yml up -d
```

## 故障排除

### 1. 服务无法启动

```bash
# 查看详细日志
docker-compose -f docker-compose.prod.yml logs

# 检查容器状态
docker-compose -f docker-compose.prod.yml ps
```

### 2. 数据库连接问题

```bash
# 检查数据库是否健康
docker-compose -f docker-compose.prod.yml exec postgres pg_isready -U anywebsites

# 进入数据库容器
docker-compose -f docker-compose.prod.yml exec postgres psql -U anywebsites anywebsites
```

### 3. 网络问题

```bash
# 检查网络连接
docker network ls
docker network inspect anywebsites_anywebsites_network
```

## 支持

如有问题，请联系技术支持或查看项目文档。
