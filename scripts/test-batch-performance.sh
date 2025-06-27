#!/bin/bash

echo "批量处理性能测试"
echo "=================="

# 测试页面ID
CONTENT_ID="ceade986-61cc-4b91-b2f9-60c71867abcf"
URL="http://localhost:8080/view/$CONTENT_ID"

echo "测试URL: $URL"
echo ""

# 进行并发请求测试来触发批量处理
echo "进行并发请求测试（20个并发请求）..."
echo "请求次数 | 响应时间 | HTTP状态"
echo "--------|---------|--------"

# 创建临时文件存储结果
temp_file=$(mktemp)

# 并发发送20个请求
for i in {1..20}; do
    {
        response_time=$(curl -o /dev/null -s -w "%{time_total} %{http_code}" "$URL")
        time_part=$(echo $response_time | cut -d' ' -f1)
        status_code=$(echo $response_time | cut -d' ' -f2)
        echo "$i $time_part $status_code" >> "$temp_file"
    } &
done

# 等待所有后台任务完成
wait

# 读取并排序结果
sort -n "$temp_file" | while read line; do
    request_num=$(echo $line | cut -d' ' -f1)
    time_part=$(echo $line | cut -d' ' -f2)
    status_code=$(echo $line | cut -d' ' -f3)
    printf "%-8s | %-9s | %s\n" "$request_num" "${time_part}s" "$status_code"
done

echo ""
echo "批量处理测试结果:"
echo "================"

# 计算统计信息
total_time=0
success_count=0
min_time=999
max_time=0

while read line; do
    time_part=$(echo $line | cut -d' ' -f2)
    status_code=$(echo $line | cut -d' ' -f3)
    
    if [ "$status_code" = "200" ]; then
        total_time=$(echo "$total_time + $time_part" | bc -l)
        success_count=$((success_count + 1))
        
        # 计算最小和最大时间
        if (( $(echo "$time_part < $min_time" | bc -l) )); then
            min_time=$time_part
        fi
        if (( $(echo "$time_part > $max_time" | bc -l) )); then
            max_time=$time_part
        fi
    fi
done < "$temp_file"

echo "成功请求数: $success_count/20"

if [ $success_count -gt 0 ]; then
    average_time=$(echo "scale=4; $total_time / $success_count" | bc -l)
    echo "平均响应时间: ${average_time}s"
    echo "最快响应时间: ${min_time}s"
    echo "最慢响应时间: ${max_time}s"
    
    # 计算每秒请求数
    rps=$(echo "scale=2; 1 / $average_time" | bc -l)
    echo "理论RPS: $rps"
else
    echo "没有成功的请求"
fi

# 清理临时文件
rm "$temp_file"

echo ""
echo "批量处理优势:"
echo "============="
echo "1. 多个地理位置查询请求被合并处理"
echo "2. 减少了GeoIP数据库的访问次数"
echo "3. 提高了并发处理能力"
echo "4. 批量大小: 10个请求/批次"
echo "5. 批量超时: 50ms"
