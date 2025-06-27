#!/bin/bash

echo "异步处理性能测试"
echo "=================="

# 测试页面ID
CONTENT_ID="ceade986-61cc-4b91-b2f9-60c71867abcf"
URL="http://localhost:8080/view/$CONTENT_ID"

echo "测试URL: $URL"
echo ""

# 进行多次请求测试
echo "进行10次连续请求测试..."
echo "请求次数 | 响应时间 | HTTP状态"
echo "--------|---------|--------"

total_time=0
success_count=0

for i in {1..10}; do
    # 使用curl测量响应时间
    response=$(curl -o /dev/null -s -w "%{time_total} %{http_code}" "$URL")

    # 提取时间和状态码
    time_part=$(echo $response | cut -d' ' -f1)
    status_code=$(echo $response | cut -d' ' -f2)
    
    printf "%-8d | %-9s | %s\n" "$i" "${time_part}s" "$status_code"
    
    # 累计成功的请求时间
    if [ "$status_code" = "200" ]; then
        total_time=$(echo "$total_time + $time_part" | bc -l)
        success_count=$((success_count + 1))
    fi
    
    # 短暂延迟避免过于频繁的请求
    sleep 0.1
done

echo ""
echo "测试结果统计:"
echo "============="
echo "成功请求数: $success_count/10"

if [ $success_count -gt 0 ]; then
    average_time=$(echo "scale=4; $total_time / $success_count" | bc -l)
    echo "平均响应时间: ${average_time}s"
    
    # 计算每秒请求数
    rps=$(echo "scale=2; 1 / $average_time" | bc -l)
    echo "理论RPS: $rps"
else
    echo "没有成功的请求"
fi

echo ""
echo "异步处理优势:"
echo "============="
echo "1. 页面响应不会被GeoIP查询阻塞"
echo "2. 用户体验更好，页面加载更快"
echo "3. 即使GeoIP查询失败，页面仍能正常显示"
echo "4. 地理位置数据在后台异步记录"
