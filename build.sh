#!/bin/bash

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

# 获取脚本所在目录
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$SCRIPT_DIR" || exit 1

# 创建输出目录
mkdir -p output/bin output/config

# 设置 Go 环境变量
export GOOS=linux    # 目标操作系统
export GOARCH=amd64  # 目标架构
export CGO_ENABLED=0 # 禁用 CGO，保证更好的兼容性

# 编译二进制
info "开始编译..."
go build -o output/bin/transfer cmd/main.go || error "编译失败"

# 复制配置文件
info "复制配置文件..."
cp -R config output/ || error "复制配置文件失败"

# 复制服务脚本
info "复制服务脚本..."
cp scripts/service.sh output/ || error "复制服务脚本失败"
chmod +x output/service.sh

# 显示编译结果
info "编译成功!"
info "编译产物:"
echo "二进制文件:"
ls -lh output/bin/
echo "配置文件:"
ls -lh output/config/

exit 0

