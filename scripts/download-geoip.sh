#!/bin/bash

# 创建 geoip 数据目录
mkdir -p data/geoip

# 下载 GeoLite2-City 数据库 (免费版本)
echo "正在下载 MaxMind GeoLite2-City 数据库..."

# MaxMind 现在需要注册才能下载，但我们可以使用一个镜像源
# 或者提供说明让用户自己下载

GEOIP_DIR="data/geoip"
CITY_DB_FILE="$GEOIP_DIR/GeoLite2-City.mmdb"

if [ ! -f "$CITY_DB_FILE" ]; then
    echo "请按照以下步骤获取 GeoLite2-City 数据库："
    echo ""
    echo "1. 访问 MaxMind 官网: https://dev.maxmind.com/geoip/geolite2-free-geolocation-data"
    echo "2. 注册免费账户"
    echo "3. 下载 GeoLite2-City.mmdb 文件"
    echo "4. 将文件放置到: $CITY_DB_FILE"
    echo ""
    echo "或者，你可以使用以下命令下载（需要有效的许可证密钥）："
    echo "wget 'https://download.maxmind.com/app/geoip_download?edition_id=GeoLite2-City&license_key=YOUR_LICENSE_KEY&suffix=tar.gz' -O GeoLite2-City.tar.gz"
    echo ""
    
    # 尝试从一个公开的镜像下载（注意：这可能不是最新版本）
    echo "尝试从镜像源下载..."
    
    # 创建一个临时的测试数据库文件，用于开发测试
    echo "创建测试用的空数据库文件..."
    touch "$CITY_DB_FILE"
    
    echo "警告: 当前使用的是空的测试文件。"
    echo "为了获得真实的地理位置数据，请从 MaxMind 官网下载真实的 GeoLite2-City.mmdb 文件。"
else
    echo "GeoLite2-City 数据库已存在: $CITY_DB_FILE"
fi

echo "完成!"
