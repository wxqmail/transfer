#!/bin/bash

# 配置信息
REMOTE_USER="ubuntu"
REMOTE_PATH="/home/ubuntu"
DEPLOY_PATH="/home/ubuntu/transfer"
PACKAGE_NAME="transfer.tar.gz"
TMP_DEPLOY_PATH="/tmp/transfer_deploy_tmp"

# 环境配置
TEST_HOSTS="43.159.63.155"  # test环境机器列表
ECS_HOSTS="39.97.49.20"  # ECS环境机器列表 - 请替换为实际IP

# 颜色输出
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

# 打印信息
info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

# 打印命令
cmd() {
    echo -e "${GREEN}[CMD]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
    exit 1
}

# 1. 编译和打包
build() {
    info "开始编译..."
    cmd "make clean"
    make clean >/dev/null
    cmd "make build"
    make build >/dev/null || error "编译失败"
    
    # 如果文件存在则删除
    if [ -f "$PACKAGE_NAME" ]; then
        cmd "rm $PACKAGE_NAME"
        rm "$PACKAGE_NAME"
    fi
    
    info "打包文件..."
    cmd "cd output && COPYFILE_DISABLE=1 tar --no-xattrs -czf ../$PACKAGE_NAME *"
    (cd output && COPYFILE_DISABLE=1 tar --no-xattrs -czf "../$PACKAGE_NAME" * 2>/dev/null) || error "打包失败"
    
    echo "$PACKAGE_NAME"
}

# 2. 上传到指定主机
upload() {
    local host=$1
    info "上传文件到 $host..."
    
    # 确保远程目录存在
    cmd "ssh $REMOTE_USER@$host mkdir -p $REMOTE_PATH"
    ssh "$REMOTE_USER@$host" "mkdir -p $REMOTE_PATH" 2>/dev/null
    
    # 上传文件
    cmd "scp $PACKAGE_NAME $REMOTE_USER@$host:$REMOTE_PATH/"
    scp "$PACKAGE_NAME" "$REMOTE_USER@$host:$REMOTE_PATH/" >/dev/null 2>&1 || error "上传到 $host 失败"
}

# 3. 在指定主机上部署
deploy() {
    local host=$1
    local env=$2
    info "在 $host 上部署..."
    
    # 构建部署命令
    local deploy_cmd=""
    
    deploy_cmd="cd $REMOTE_PATH && \
        rm -rf $TMP_DEPLOY_PATH && mkdir -p $TMP_DEPLOY_PATH && \
        tar -xzf $PACKAGE_NAME -C $TMP_DEPLOY_PATH && \
        mkdir -p $DEPLOY_PATH && \
        rm -rf $DEPLOY_PATH/bin $DEPLOY_PATH/config $DEPLOY_PATH/service.sh && \
        cd $TMP_DEPLOY_PATH && \
        cp -r bin config service.sh $DEPLOY_PATH/ && \
        cd $DEPLOY_PATH && chmod +x service.sh && \
        echo '开始重启服务...' && \
        ./service.sh restart 2>&1 && \
        cd $REMOTE_PATH && rm -rf $TMP_DEPLOY_PATH"
    
    cmd "ssh $REMOTE_USER@$host '$deploy_cmd'"
    ssh "$REMOTE_USER@$host" "$deploy_cmd" || error "在 $host 上部署失败"
    
    # 清理远程包
    cmd "ssh $REMOTE_USER@$host 'cd $REMOTE_PATH && rm -f $PACKAGE_NAME'"
    ssh "$REMOTE_USER@$host" "cd $REMOTE_PATH && rm -f $PACKAGE_NAME" >/dev/null 2>&1
}

# 显示使用方法
usage() {
    echo "Usage: $0 <env>"
    echo "env: test|ecs"
    exit 1
}

# 主流程
main() {
    # 检查参数
    if [ $# -ne 1 ]; then
        usage
    fi

    local env=$1
    local hosts=""

    # 根据环境选择主机列表
    case $env in
        "test")
            hosts=$TEST_HOSTS
            ;;
        "ecs")
            hosts=$ECS_HOSTS
            ;;
        *)
            error "未知的环境: $env"
            ;;
    esac

    info "开始部署到 $env 环境..."

    # 构建
    build

    # 对每个主机进行部署
    for host in $hosts; do
        info "开始处理主机: $host"
        upload "$host"
        deploy "$host" "$env"
        info "主机 $host 部署完成"
    done

    info "所有机器部署完成!"
}

main "$@"