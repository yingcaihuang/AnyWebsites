FROM nginx:alpine

# 安装 openssl 用于生成自签名证书
RUN apk add --no-cache openssl

# 复制 nginx 配置文件
COPY nginx.conf /etc/nginx/nginx.conf

# 复制 SSL 证书生成脚本
COPY generate-ssl.sh /usr/local/bin/generate-ssl.sh
RUN chmod +x /usr/local/bin/generate-ssl.sh

# 创建启动脚本
RUN echo '#!/bin/sh' > /docker-entrypoint.sh && \
    echo 'set -e' >> /docker-entrypoint.sh && \
    echo '' >> /docker-entrypoint.sh && \
    echo '# 生成 SSL 证书（如果不存在）' >> /docker-entrypoint.sh && \
    echo 'if [ ! -f /etc/nginx/ssl/cert.pem ]; then' >> /docker-entrypoint.sh && \
    echo '    echo "Generating SSL certificates..."' >> /docker-entrypoint.sh && \
    echo '    mkdir -p /etc/nginx/ssl' >> /docker-entrypoint.sh && \
    echo '    openssl req -x509 -nodes -days 365 -newkey rsa:2048 \' >> /docker-entrypoint.sh && \
    echo '        -keyout /etc/nginx/ssl/key.pem \' >> /docker-entrypoint.sh && \
    echo '        -out /etc/nginx/ssl/cert.pem \' >> /docker-entrypoint.sh && \
    echo '        -subj "/C=CN/ST=Beijing/L=Beijing/O=AnyWebsites/OU=IT/CN=localhost"' >> /docker-entrypoint.sh && \
    echo '    chmod 600 /etc/nginx/ssl/key.pem' >> /docker-entrypoint.sh && \
    echo '    chmod 644 /etc/nginx/ssl/cert.pem' >> /docker-entrypoint.sh && \
    echo 'fi' >> /docker-entrypoint.sh && \
    echo '' >> /docker-entrypoint.sh && \
    echo '# 启动 nginx' >> /docker-entrypoint.sh && \
    echo 'exec nginx -g "daemon off;"' >> /docker-entrypoint.sh && \
    chmod +x /docker-entrypoint.sh

# 暴露端口
EXPOSE 80 443

# 启动脚本
ENTRYPOINT ["/docker-entrypoint.sh"]
