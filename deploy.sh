#!/bin/bash

# AnyWebsites 部署脚本
# 使用方法: ./deploy.sh [start|stop|restart|status|logs|update]

set -e

COMPOSE_FILE="docker-compose.prod.yml"
PROJECT_NAME="anywebsites"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查Docker和Docker Compose
check_requirements() {
    log_info "检查系统要求..."
    
    if ! command -v docker &> /dev/null; then
        log_error "Docker 未安装，请先安装 Docker"
        exit 1
    fi
    
    if ! command -v docker-compose &> /dev/null; then
        log_error "Docker Compose 未安装，请先安装 Docker Compose"
        exit 1
    fi
    
    log_success "系统要求检查通过"
}

# 检查必要文件
check_files() {
    log_info "检查必要文件..."
    
    if [ ! -f "$COMPOSE_FILE" ]; then
        log_error "找不到 $COMPOSE_FILE 文件"
        exit 1
    fi
    
    if [ ! -f ".env" ]; then
        log_warning "找不到 .env 文件，将使用默认配置"
    fi
    
    if [ ! -f "init.sql" ]; then
        log_warning "找不到 init.sql 文件，数据库可能无法正确初始化"
    fi
    
    log_success "文件检查完成"
}

# 拉取最新镜像
pull_images() {
    log_info "拉取最新镜像..."
    docker-compose -f $COMPOSE_FILE pull
    log_success "镜像拉取完成"
}

# 启动服务
start_services() {
    log_info "启动服务..."
    docker-compose -f $COMPOSE_FILE up -d
    log_success "服务启动完成"
}

# 停止服务
stop_services() {
    log_info "停止服务..."
    docker-compose -f $COMPOSE_FILE down
    log_success "服务停止完成"
}

# 重启服务
restart_services() {
    log_info "重启服务..."
    docker-compose -f $COMPOSE_FILE restart
    log_success "服务重启完成"
}

# 查看服务状态
show_status() {
    log_info "服务状态:"
    docker-compose -f $COMPOSE_FILE ps
}

# 查看日志
show_logs() {
    log_info "显示服务日志..."
    docker-compose -f $COMPOSE_FILE logs -f --tail=100
}

# 更新部署
update_deployment() {
    log_info "更新部署..."
    pull_images
    docker-compose -f $COMPOSE_FILE down
    start_services
    log_success "部署更新完成"
}

# 健康检查
health_check() {
    log_info "执行健康检查..."
    
    # 等待服务启动
    sleep 10
    
    # 检查HTTP服务
    if curl -f -s http://localhost > /dev/null; then
        log_success "HTTP 服务正常"
    else
        log_warning "HTTP 服务可能未正常启动"
    fi
    
    # 检查HTTPS服务
    if curl -f -s -k https://localhost > /dev/null; then
        log_success "HTTPS 服务正常"
    else
        log_warning "HTTPS 服务可能未正常启动"
    fi
}

# 显示帮助信息
show_help() {
    echo "AnyWebsites 部署脚本"
    echo ""
    echo "使用方法:"
    echo "  $0 [命令]"
    echo ""
    echo "可用命令:"
    echo "  start     启动所有服务"
    echo "  stop      停止所有服务"
    echo "  restart   重启所有服务"
    echo "  status    查看服务状态"
    echo "  logs      查看服务日志"
    echo "  update    更新部署（拉取最新镜像并重启）"
    echo "  health    执行健康检查"
    echo "  help      显示此帮助信息"
    echo ""
    echo "示例:"
    echo "  $0 start    # 启动服务"
    echo "  $0 logs     # 查看日志"
    echo "  $0 update   # 更新部署"
}

# 主函数
main() {
    case "${1:-help}" in
        start)
            check_requirements
            check_files
            pull_images
            start_services
            health_check
            show_status
            ;;
        stop)
            stop_services
            ;;
        restart)
            restart_services
            health_check
            show_status
            ;;
        status)
            show_status
            ;;
        logs)
            show_logs
            ;;
        update)
            check_requirements
            update_deployment
            health_check
            show_status
            ;;
        health)
            health_check
            ;;
        help|--help|-h)
            show_help
            ;;
        *)
            log_error "未知命令: $1"
            show_help
            exit 1
            ;;
    esac
}

# 执行主函数
main "$@"
