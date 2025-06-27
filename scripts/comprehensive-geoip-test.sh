#!/bin/bash

echo "GeoIP 服务综合性能测试"
echo "======================"

# 测试配置
CONTENT_ID="ceade986-61cc-4b91-b2f9-60c71867abcf"
BASE_URL="http://localhost:8080"
VIEW_URL="$BASE_URL/view/$CONTENT_ID"
STATS_URL="$BASE_URL/admin/api/geoip-stats"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 打印带颜色的消息
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查服务器是否运行
check_server() {
    print_status "检查服务器状态..."
    if curl -s "$BASE_URL/health" > /dev/null; then
        print_success "服务器运行正常"
        return 0
    else
        print_error "服务器未运行或无法访问"
        return 1
    fi
}

# 获取初始统计信息
get_initial_stats() {
    print_status "获取初始统计信息..."
    
    # 尝试获取统计信息（可能需要登录）
    initial_stats=$(curl -s "$STATS_URL" 2>/dev/null)
    
    if echo "$initial_stats" | grep -q "success.*true"; then
        print_success "成功获取初始统计信息"
        echo "$initial_stats" | jq '.' 2>/dev/null || echo "$initial_stats"
    else
        print_warning "无法获取详细统计信息（可能需要管理员登录）"
        echo "将继续进行基本测试..."
    fi
    echo ""
}

# 单次请求测试
single_request_test() {
    print_status "执行单次请求测试..."
    
    start_time=$(date +%s.%N)
    response=$(curl -s -w "%{http_code}" -o /dev/null "$VIEW_URL")
    end_time=$(date +%s.%N)
    
    duration=$(echo "$end_time - $start_time" | bc -l)
    
    if [ "$response" = "200" ]; then
        print_success "单次请求成功 - 响应时间: ${duration}s"
    else
        print_error "单次请求失败 - HTTP状态码: $response"
        return 1
    fi
}

# 并发请求测试
concurrent_test() {
    local concurrent_count=$1
    print_status "执行并发测试 ($concurrent_count 个并发请求)..."
    
    # 创建临时文件存储结果
    temp_file=$(mktemp)
    
    # 记录开始时间
    start_time=$(date +%s.%N)
    
    # 并发发送请求
    for i in $(seq 1 $concurrent_count); do
        {
            request_start=$(date +%s.%N)
            response_code=$(curl -s -w "%{http_code}" -o /dev/null "$VIEW_URL")
            request_end=$(date +%s.%N)
            request_duration=$(echo "$request_end - $request_start" | bc -l)
            echo "$i $response_code $request_duration" >> "$temp_file"
        } &
    done
    
    # 等待所有请求完成
    wait
    
    # 记录结束时间
    end_time=$(date +%s.%N)
    total_duration=$(echo "$end_time - $start_time" | bc -l)
    
    # 分析结果
    success_count=0
    total_response_time=0
    min_time=999
    max_time=0
    
    while read line; do
        request_num=$(echo $line | cut -d' ' -f1)
        status_code=$(echo $line | cut -d' ' -f2)
        response_time=$(echo $line | cut -d' ' -f3)
        
        if [ "$status_code" = "200" ]; then
            success_count=$((success_count + 1))
            total_response_time=$(echo "$total_response_time + $response_time" | bc -l)
            
            # 更新最小和最大时间
            if (( $(echo "$response_time < $min_time" | bc -l) )); then
                min_time=$response_time
            fi
            if (( $(echo "$response_time > $max_time" | bc -l) )); then
                max_time=$response_time
            fi
        fi
    done < "$temp_file"
    
    # 计算统计信息
    if [ $success_count -gt 0 ]; then
        avg_response_time=$(echo "scale=4; $total_response_time / $success_count" | bc -l)
        throughput=$(echo "scale=2; $concurrent_count / $total_duration" | bc -l)
        
        print_success "并发测试完成"
        echo "  成功请求数: $success_count/$concurrent_count"
        echo "  总耗时: ${total_duration}s"
        echo "  平均响应时间: ${avg_response_time}s"
        echo "  最快响应时间: ${min_time}s"
        echo "  最慢响应时间: ${max_time}s"
        echo "  吞吐量: ${throughput} 请求/秒"
    else
        print_error "所有并发请求都失败了"
    fi
    
    # 清理临时文件
    rm "$temp_file"
    echo ""
}

# 缓存性能测试
cache_performance_test() {
    print_status "执行缓存性能测试..."
    
    # 第一次请求（缓存未命中）
    print_status "第一次请求（预期缓存未命中）..."
    start_time=$(date +%s.%N)
    curl -s "$VIEW_URL" > /dev/null
    end_time=$(date +%s.%N)
    first_request_time=$(echo "$end_time - $start_time" | bc -l)
    
    # 等待一秒确保地理位置处理完成
    sleep 1
    
    # 第二次请求（预期缓存命中）
    print_status "第二次请求（预期缓存命中）..."
    start_time=$(date +%s.%N)
    curl -s "$VIEW_URL" > /dev/null
    end_time=$(date +%s.%N)
    second_request_time=$(echo "$end_time - $start_time" | bc -l)
    
    # 比较性能
    if (( $(echo "$second_request_time < $first_request_time" | bc -l) )); then
        improvement=$(echo "scale=2; ($first_request_time - $second_request_time) / $first_request_time * 100" | bc -l)
        print_success "缓存性能提升明显"
        echo "  第一次请求: ${first_request_time}s"
        echo "  第二次请求: ${second_request_time}s"
        echo "  性能提升: ${improvement}%"
    else
        print_warning "缓存性能提升不明显或测试环境影响"
        echo "  第一次请求: ${first_request_time}s"
        echo "  第二次请求: ${second_request_time}s"
    fi
    echo ""
}

# 批量处理测试
batch_processing_test() {
    print_status "执行批量处理测试..."
    
    # 快速连续发送多个请求来触发批量处理
    print_status "发送快速连续请求来触发批量处理..."
    
    for i in {1..15}; do
        curl -s "$VIEW_URL" > /dev/null &
    done
    
    # 等待所有请求完成
    wait
    
    print_success "批量处理测试完成"
    print_status "检查服务器日志以确认批量处理是否被触发"
    echo ""
}

# 压力测试
stress_test() {
    print_status "执行压力测试..."
    
    # 高并发测试
    concurrent_test 50
    
    # 持续负载测试
    print_status "执行持续负载测试（30秒）..."
    
    end_time=$(($(date +%s) + 30))
    request_count=0
    success_count=0
    
    while [ $(date +%s) -lt $end_time ]; do
        if curl -s "$VIEW_URL" > /dev/null; then
            success_count=$((success_count + 1))
        fi
        request_count=$((request_count + 1))
        sleep 0.1  # 100ms间隔
    done
    
    success_rate=$(echo "scale=2; $success_count * 100 / $request_count" | bc -l)
    rps=$(echo "scale=2; $request_count / 30" | bc -l)
    
    print_success "持续负载测试完成"
    echo "  总请求数: $request_count"
    echo "  成功请求数: $success_count"
    echo "  成功率: ${success_rate}%"
    echo "  平均RPS: $rps"
    echo ""
}

# 主测试流程
main() {
    echo "开始时间: $(date)"
    echo ""
    
    # 检查服务器
    if ! check_server; then
        exit 1
    fi
    echo ""
    
    # 获取初始统计
    get_initial_stats
    
    # 执行各种测试
    single_request_test
    echo ""
    
    concurrent_test 5
    concurrent_test 10
    concurrent_test 20
    
    cache_performance_test
    batch_processing_test
    stress_test
    
    # 最终统计
    print_status "获取最终统计信息..."
    final_stats=$(curl -s "$STATS_URL" 2>/dev/null)
    
    if echo "$final_stats" | grep -q "success.*true"; then
        print_success "最终统计信息:"
        echo "$final_stats" | jq '.' 2>/dev/null || echo "$final_stats"
    else
        print_warning "无法获取最终统计信息"
    fi
    
    echo ""
    echo "测试完成时间: $(date)"
    print_success "所有测试已完成！"
}

# 运行主程序
main
