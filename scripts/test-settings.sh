#!/bin/bash

# 系统设置功能测试脚本
echo "系统设置功能测试"
echo "=================="

# 服务器地址
SERVER="http://localhost:8080"
ADMIN_API="$SERVER/admin/api"

# 管理员登录获取 Cookie
echo "[INFO] 正在登录管理员账户..."
LOGIN_RESPONSE=$(curl -s -c cookies.txt -X POST "$SERVER/admin/login" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "username=admin&password=admin123")

if [[ $LOGIN_RESPONSE == *"重定向"* ]] || [[ $LOGIN_RESPONSE == *"dashboard"* ]]; then
    echo "[SUCCESS] 管理员登录成功"
else
    echo "[ERROR] 管理员登录失败"
    echo "响应: $LOGIN_RESPONSE"
    exit 1
fi

# 测试获取所有分类
echo ""
echo "[INFO] 测试获取设置分类..."
CATEGORIES_RESPONSE=$(curl -s -b cookies.txt "$ADMIN_API/settings/categories")
echo "分类响应: $CATEGORIES_RESPONSE"

# 测试获取服务器分类的设置
echo ""
echo "[INFO] 测试获取服务器分类设置..."
SERVER_SETTINGS_RESPONSE=$(curl -s -b cookies.txt "$ADMIN_API/settings/category/server")
echo "服务器设置响应: $SERVER_SETTINGS_RESPONSE"

# 测试创建一个新设置
echo ""
echo "[INFO] 测试创建新设置..."
CREATE_RESPONSE=$(curl -s -b cookies.txt -X POST "$ADMIN_API/settings" \
  -H "Content-Type: application/json" \
  -d '{
    "category": "server",
    "key": "test_port",
    "value": 8080,
    "description": "测试端口设置",
    "reason": "自动化测试创建"
  }')
echo "创建设置响应: $CREATE_RESPONSE"

# 测试获取所有设置
echo ""
echo "[INFO] 测试获取所有设置..."
ALL_SETTINGS_RESPONSE=$(curl -s -b cookies.txt "$ADMIN_API/settings")
echo "所有设置响应: $ALL_SETTINGS_RESPONSE"

# 测试导出设置
echo ""
echo "[INFO] 测试导出设置..."
EXPORT_RESPONSE=$(curl -s -b cookies.txt "$ADMIN_API/settings/export")
echo "导出响应: $EXPORT_RESPONSE"

# 将导出的设置保存到文件
echo "$EXPORT_RESPONSE" > settings_backup.json
echo "[INFO] 设置已导出到 settings_backup.json"

# 测试创建另一个设置用于测试导入
echo ""
echo "[INFO] 创建另一个测试设置..."
CREATE_RESPONSE2=$(curl -s -b cookies.txt -X POST "$ADMIN_API/settings" \
  -H "Content-Type: application/json" \
  -d '{
    "category": "security",
    "key": "test_timeout",
    "value": 3600,
    "description": "测试超时设置",
    "reason": "自动化测试创建"
  }')
echo "创建第二个设置响应: $CREATE_RESPONSE2"

# 测试获取设置历史
echo ""
echo "[INFO] 测试获取设置历史..."
HISTORY_RESPONSE=$(curl -s -b cookies.txt "$ADMIN_API/settings/server/test_port/history")
echo "设置历史响应: $HISTORY_RESPONSE"

# 测试更新设置
echo ""
echo "[INFO] 测试更新设置..."
# 首先获取设置ID
SETTING_ID=$(echo "$ALL_SETTINGS_RESPONSE" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
if [ -n "$SETTING_ID" ]; then
    UPDATE_RESPONSE=$(curl -s -b cookies.txt -X PUT "$ADMIN_API/settings/$SETTING_ID" \
      -H "Content-Type: application/json" \
      -d '{
        "category": "server",
        "key": "test_port",
        "value": 9090,
        "description": "更新后的测试端口设置",
        "reason": "自动化测试更新"
      }')
    echo "更新设置响应: $UPDATE_RESPONSE"
else
    echo "[WARNING] 无法获取设置ID，跳过更新测试"
fi

# 测试删除设置
echo ""
echo "[INFO] 测试删除设置..."
if [ -n "$SETTING_ID" ]; then
    DELETE_RESPONSE=$(curl -s -b cookies.txt -X DELETE "$ADMIN_API/settings/$SETTING_ID" \
      -H "Content-Type: application/json" \
      -d '{
        "reason": "自动化测试删除"
      }')
    echo "删除设置响应: $DELETE_RESPONSE"
else
    echo "[WARNING] 无法获取设置ID，跳过删除测试"
fi

# 测试导入设置
echo ""
echo "[INFO] 测试导入设置..."
if [ -f "settings_backup.json" ]; then
    # 创建一个简单的导入测试数据
    IMPORT_DATA='{
      "backup": {
        "version": "1.0",
        "timestamp": "'$(date -u +%Y-%m-%dT%H:%M:%S.%3NZ)'",
        "categories": [],
        "settings": {
          "server.import_test": {
            "category": "server",
            "key": "import_test",
            "value": "imported_value",
            "value_type": "string",
            "description": "通过导入功能创建的测试设置"
          }
        },
        "metadata": {
          "total_settings": 1,
          "total_categories": 0,
          "exported_by": "test_script"
        }
      },
      "overwrite": true
    }'
    
    IMPORT_RESPONSE=$(curl -s -b cookies.txt -X POST "$ADMIN_API/settings/import" \
      -H "Content-Type: application/json" \
      -d "$IMPORT_DATA")
    echo "导入设置响应: $IMPORT_RESPONSE"
else
    echo "[WARNING] 备份文件不存在，跳过导入测试"
fi

# 最终验证
echo ""
echo "[INFO] 最终验证 - 获取所有设置..."
FINAL_SETTINGS_RESPONSE=$(curl -s -b cookies.txt "$ADMIN_API/settings")
echo "最终设置列表: $FINAL_SETTINGS_RESPONSE"

echo ""
echo "测试完成！"
echo "============"

# 清理临时文件
rm -f cookies.txt settings_backup.json

echo "[INFO] 临时文件已清理"
