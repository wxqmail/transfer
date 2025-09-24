#!/bin/bash

# 获取脚本所在目录的绝对路径
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

# 服务配置
SERVICE_NAME="transfer"
PID_FILE="logs/transfer.pid"
LOG_FILE="logs/transfer.log"
BIN_PATH="bin/transfer"

# 确保在正确的目录下执行
cd "$SCRIPT_DIR" || {
    echo "无法切换到脚本目录: $SCRIPT_DIR"
    exit 1
}

# 创建日志目录
mkdir -p "$(dirname "$PID_FILE")" "$(dirname "$LOG_FILE")"

# 颜色输出
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

# 打印信息
info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
    exit 1
}

# 获取服务PID
get_pid() {
    if [ -f "$PID_FILE" ]; then
        cat "$PID_FILE"
    fi
}

# 通过进程名获取PID
get_pid_by_name() {
    pgrep -f "$BIN_PATH"
}

# 检查服务状态
check_status() {
    pid=$(get_pid)
    if [ -n "$pid" ]; then
        if ps -p "$pid" >/dev/null 2>&1; then
            return 0
        fi
    fi
    return 1
}

# 启动服务
start() {
    info "正在启动 $SERVICE_NAME 服务..."
    
    # 检查二进制文件是否存在
    if [ ! -f "$BIN_PATH" ]; then
        error "服务程序不存在: $BIN_PATH"
    fi

    if check_status; then
        error "服务已经在运行中 (PID: $(get_pid))"
    fi
    
    echo "正在启动 $SERVICE_NAME..."
    nohup "./$BIN_PATH" > "$LOG_FILE" 2>&1 & echo $! > "$PID_FILE"
    sleep 1
    if check_status; then
        info "服务启动成功 (PID: $(get_pid))"
    else
        error "服务启动失败"
    fi
}

# 停止服务
stop() {
    info "正在停止 $SERVICE_NAME 服务..."
    
    # 获取PID文件中的PID
    pid=$(get_pid)
    
    # 获取通过进程名找到的PID
    name_pid=$(get_pid_by_name)
    
    # 如果PID文件存在，尝试停止对应进程
    if [ -n "$pid" ]; then
        kill "$pid" 2>/dev/null
        rm -f "$PID_FILE"
    fi
    
    # 如果通过进程名找到了PID，也尝试停止
    if [ -n "$name_pid" ]; then
        for p in $name_pid; do
            if [ "$p" != "$pid" ]; then
                kill "$p" 2>/dev/null
            fi
        done
    fi
    
    # 等待进程停止
    sleep 2
    
    # 检查是否还有进程在运行
    name_pid=$(get_pid_by_name)
    if [ -n "$name_pid" ]; then
        for p in $name_pid; do
            kill -9 "$p" 2>/dev/null
        done
        sleep 1
    fi
    
    # 确保PID文件被删除
    rm -f "$PID_FILE"
    
    # 再次检查是否有进程在运行
    if [ -n "$(get_pid_by_name)" ]; then
        error "服务停止失败"
        return 1
    else
        info "服务已停止"
        return 0
    fi
}

# 重启服务
restart() {
    info "正在重启 $SERVICE_NAME 服务..."
    stop
    sleep 1
    start
}

# 查看状态
status() {
    if check_status; then
        info "服务运行中 (PID: $(get_pid))"
        echo "端口状态:"
        netstat -tlnp | grep :8199 || echo "端口8199未监听"
    else
        info "服务未运行"
    fi
}

# 命令处理
case "$1" in
    start)
        start
        ;;
    stop)
        stop
        ;;
    restart)
        restart
        ;;
    status)
        status
        ;;
    *)
        echo "用法: $0 {start|stop|restart|status}"
        exit 1
        ;;
esac

exit 0
