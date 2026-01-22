#!/bin/bash

# Claude Switcher 一键安装脚本
# 用法: curl -fsSL https://raw.githubusercontent.com/fiftyk/claude-switcher/main/install.sh | bash

set -e

# 颜色定义
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

echo_info() {
    echo -e "${YELLOW}ℹ $1${NC}"
}

echo_error() {
    echo -e "${RED}✗ $1${NC}" >&2
}

# 配置变量
REPO="fiftyk/claude-switcher"
BINARY_NAME="claude-switcher"
INSTALL_DIR="/usr/local/bin"

# 检测操作系统
OS=$(uname -s)
ARCH=$(uname -m)

# 将架构映射到 Go 格式
case "$OS" in
    Darwin)
        GO_OS="darwin"
        ;;
    Linux)
        GO_OS="linux"
        ;;
    *)
        echo_error "不支持的操作系统: $OS"
        exit 1
        ;;
esac

case "$ARCH" in
    x86_64)
        GO_ARCH="amd64"
        ;;
    arm64|aarch64)
        GO_ARCH="arm64"
        ;;
    *)
        echo_error "不支持的架构: $ARCH"
        exit 1
        ;;
esac

echo "=== Claude Switcher 安装脚本 ==="
echo "检测到系统: $OS ($ARCH)"
echo ""

# 获取最新版本
echo_info "获取最新版本..."
LATEST_TAG=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" 2>/dev/null | grep '"tag_name"' | sed 's/.*": "\(.*\)".*/\1/' || echo "")

if [ -z "$LATEST_TAG" ]; then
    echo_info "无法获取最新版本，使用默认版本"
    LATEST_TAG="latest"
fi

echo "版本: $LATEST_TAG"

# 下载链接
if [ "$LATEST_TAG" = "latest" ]; then
    DOWNLOAD_URL="https://github.com/${REPO}/releases/latest/download/${BINARY_NAME}-${GO_OS}-${GO_ARCH}"
else
    DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${LATEST_TAG}/${BINARY_NAME}-${GO_OS}-${GO_ARCH}"
fi

# 安装目录
INSTALL_PATH="${INSTALL_DIR}/${BINARY_NAME}"

echo ""
echo_info "下载安装包..."
echo "URL: $DOWNLOAD_URL"

# 创建临时文件
TEMP_FILE=$(mktemp)
trap "rm -f $TEMP_FILE" EXIT

# 下载
if ! curl -fsSL "$DOWNLOAD_URL" -o "$TEMP_FILE"; then
    echo_error "下载失败"
    exit 1
fi

# 验证下载的文件
if [ ! -s "$TEMP_FILE" ]; then
    echo_error "下载的文件为空"
    exit 1
fi

# 设置权限
chmod +x "$TEMP_FILE"

# 检查是否需要 sudo
NEED_SUDO=false
if [ ! -w "$INSTALL_DIR" ]; then
    NEED_SUDO=true
fi

# 备份现有文件（如果存在）
if [ -f "$INSTALL_PATH" ]; then
    BACKUP_FILE="${INSTALL_PATH}.backup.$(date +%s)"
    echo_info "备份现有文件到: $BACKUP_FILE"
    if [ "$NEED_SUDO" = true ]; then
        sudo cp "$INSTALL_PATH" "$BACKUP_FILE"
    else
        cp "$INSTALL_PATH" "$BACKUP_FILE"
    fi
fi

# 安装
echo_info "安装到: $INSTALL_PATH"
if [ "$NEED_SUDO" = true ]; then
    sudo mv "$TEMP_FILE" "$INSTALL_PATH"
    TEMP_FILE=""  # 已经移动，trap 不会删除
else
    mv "$TEMP_FILE" "$INSTALL_PATH"
    TEMP_FILE=""  # 已经移动，trap 不会删除
fi

# 验证安装
if [ -x "$INSTALL_PATH" ]; then
    echo ""
    echo_success "安装成功!"
    echo ""
    echo_info "使用方法:"
    echo "  claude-switcher                    # 交互式选择配置"
    echo "  claude-switcher <配置名称>         # 使用指定配置启动"
    echo "  claude-switcher <配置名称> --sync  # 切换并同步到 settings.json"
    echo "  claude-switcher --list             # 列出所有配置"
    echo "  claude-switcher --help             # 显示帮助"
    echo ""
    echo "更多信息: https://github.com/${REPO}"
else
    echo_error "安装验证失败"
    exit 1
fi
