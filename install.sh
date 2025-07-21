#!/bin/bash

# Claude Switcher 简化安装脚本

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

# 安装目录
INSTALL_DIR="/usr/local/bin"
SCRIPT_NAME="claude-switcher"

echo_info "安装 Claude Switcher..."

# 检查是否安装过旧版本 claude-proxy-checker
check_old_version() {
    local old_script="$INSTALL_DIR/claude-proxy-checker"
    local old_config="$HOME/.claude_proxy_config"
    
    if [ -f "$old_script" ] || [ -f "$old_config" ]; then
        echo_info "检测到旧版本 claude-proxy-checker"
        echo_info "建议删除旧版本以避免冲突"
        
        echo -n "是否删除旧版本？[y/N]: "
        read -r remove_old < /dev/tty
        
        if [[ "$remove_old" =~ ^[yY] ]]; then
            # 删除旧脚本
            if [ -f "$old_script" ]; then
                if [ -w "$INSTALL_DIR" ]; then
                    rm -f "$old_script"
                else
                    sudo rm -f "$old_script"
                fi
                echo_success "已删除旧脚本: $old_script"
            fi
            
            # 删除旧配置
            if [ -f "$old_config" ]; then
                echo -n "是否同时删除旧配置文件？[y/N]: "
                read -r remove_config < /dev/tty
                
                if [[ "$remove_config" =~ ^[yY] ]]; then
                    rm -f "$old_config"
                    echo_success "已删除旧配置: $old_config"
                else
                    echo_info "保留旧配置文件: $old_config"
                fi
            fi
        else
            echo_info "保留旧版本，请注意可能的命令冲突"
        fi
    fi
}

# 检查旧版本
check_old_version

# 检查权限
if [ ! -w "$INSTALL_DIR" ]; then
    echo_info "需要管理员权限..."
    NEED_SUDO=true
else
    NEED_SUDO=false
fi

# 下载主脚本
echo_info "下载脚本文件..."
curl -L https://raw.githubusercontent.com/fiftyk/claude-switcher/main/claude-switcher.sh -o claude-switcher.sh

# 设置权限
chmod +x claude-switcher.sh

# 安装到系统路径
if [ "$NEED_SUDO" = true ]; then
    sudo mv claude-switcher.sh "$INSTALL_DIR/$SCRIPT_NAME"
else
    mv claude-switcher.sh "$INSTALL_DIR/$SCRIPT_NAME"
fi

# 验证安装
if command -v claude-switcher &> /dev/null; then
    echo_success "Claude Switcher 安装成功！"
    echo
    echo_info "使用方法:"
    echo "  claude-switcher    # 启动程序"
    echo
    echo_info "首次运行会引导你创建配置"
else
    echo_error "安装失败"
    exit 1
fi
