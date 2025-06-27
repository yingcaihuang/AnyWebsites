#!/bin/bash

# 创建SSL证书目录
mkdir -p /etc/nginx/ssl

# 生成自签名SSL证书（用于测试）
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
    -keyout /etc/nginx/ssl/key.pem \
    -out /etc/nginx/ssl/cert.pem \
    -subj "/C=CN/ST=Beijing/L=Beijing/O=AnyWebsites/OU=IT/CN=localhost"

# 设置正确的权限
chmod 600 /etc/nginx/ssl/key.pem
chmod 644 /etc/nginx/ssl/cert.pem

echo "SSL certificates generated successfully!"
