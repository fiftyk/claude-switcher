#!/bin/bash

# Claude Switcher - 简化版本
# 智能配置管理和切换工具

# 设置颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# 配置目录
CONFIG_DIR="$HOME/.claude-switcher"
PROFILES_DIR="$CONFIG_DIR/profiles"
ACTIVE_FILE="$CONFIG_DIR/active"

# 命令行参数处理
parse_arguments() {
    local config_name=""
    
    case "$#" in
        0)
            # 无参数，使用交互式模式
            return 0
            ;;
        1)
            case "$1" in
                --help|-h)
                    show_help_info
                    exit 0
                    ;;
                --list|-l)
                    list_available_configs
                    exit 0
                    ;;
                --config)
                    echo_error "缺少配置名称"
                    echo_info "用法: claude-switcher --config <配置名称>"
                    exit 1
                    ;;
                -*)
                    echo_error "未知参数: $1"
                    show_help_info
                    exit 1
                    ;;
                *)
                    # 直接指定配置名称
                    config_name="$1"
                    ;;
            esac
            ;;
        2)
            case "$1" in
                --config|-c)
                    config_name="$2"
                    ;;
                *)
                    echo_error "未知参数组合: $1 $2"
                    show_help_info
                    exit 1
                    ;;
            esac
            ;;
        *)
            echo_error "参数过多"
            show_help_info
            exit 1
            ;;
    esac
    
    if [ -n "$config_name" ]; then
        run_with_specified_config "$config_name"
        exit 0
    fi
}

# 显示帮助信息
show_help_info() {
    echo_title "Claude Switcher - 使用帮助"
    echo
    echo -e "${YELLOW}用法:${NC}"
    echo "  claude-switcher                    启动交互式配置选择"
    echo "  claude-switcher <配置名称>         直接使用指定配置启动"
    echo "  claude-switcher --config <名称>    使用指定配置启动"
    echo "  claude-switcher --list             列出所有可用配置"
    echo "  claude-switcher --help             显示此帮助信息"
    echo
    echo -e "${YELLOW}示例:${NC}"
    echo "  claude-switcher moonshot           # 直接启动moonshot配置"
    echo "  claude-switcher --config work      # 启动work配置"
    echo "  claude-switcher --list             # 查看所有配置"
    echo
    echo -e "${YELLOW}说明:${NC}"
    echo "  • 配置文件位于: ~/.claude-switcher/profiles/"
    echo "  • 无参数运行时进入交互式菜单"
    echo "  • 指定不存在的配置会显示可用配置列表"
}

# 列出可用配置
list_available_configs() {
    echo_title "可用配置列表"
    
    local found_any=false
    
    if [ -d "$PROFILES_DIR" ]; then
        for config_file in "$PROFILES_DIR"/*.conf; do
            if [ -f "$config_file" ]; then
                found_any=true
                local name
                name=$(basename "$config_file" .conf)
                
                # 读取配置显示名称
                local display_name=""
                if [ -f "$config_file" ]; then
                    display_name=$(grep "^NAME=" "$config_file" 2>/dev/null | cut -d'=' -f2- | tr -d '"' || echo "")
                fi
                
                # 如果没有NAME字段，使用文件名
                if [ -z "$display_name" ]; then
                    display_name="$name"
                fi
                
                echo "  $name - $display_name"
            fi
        done
    fi
    
    if [ "$found_any" = false ]; then
        echo_info "暂无可用配置，请先运行 'claude-switcher' 创建配置"
    fi
}

# 使用指定配置运行
run_with_specified_config() {
    local config_name="$1"
    local config_file="$PROFILES_DIR/$config_name.conf"
    
    if [ ! -f "$config_file" ]; then
        echo_error "配置 '$config_name' 不存在"
        echo
        echo_info "可用配置:"
        list_available_configs
        exit 1
    fi
    
    # 设置为活动配置并启动
    set_active_profile "$config_name"
    echo_info "使用配置: $config_name"
    run_claude_with_profile "$config_name"
}

# 颜色输出函数
echo_error() {
    echo -e "${RED}✗ $1${NC}" >&2
}

echo_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

echo_info() {
    echo -e "${BLUE}ℹ $1${NC}"
}

echo_warning() {
    echo -e "${YELLOW}⚠ $1${NC}"
}

echo_title() {
    echo -e "\n${YELLOW}=== $1 ===${NC}"
}

# 检查依赖
check_dependencies() {
    if ! command -v curl &> /dev/null; then
        echo_error "curl 未安装，请先安装 curl"
        exit 1
    fi
}

# 初始化配置目录
init_config_dir() {
    if [ ! -d "$CONFIG_DIR" ]; then
        mkdir -p "$PROFILES_DIR"
        echo_info "配置目录初始化完成: $CONFIG_DIR"
    fi
}

# 获取当前激活的配置
get_active_profile() {
    if [ -f "$ACTIVE_FILE" ]; then
        cat "$ACTIVE_FILE"
    else
        echo ""
    fi
}

# 设置激活的配置
set_active_profile() {
    local name="$1"
    echo "$name" > "$ACTIVE_FILE"
}

# 获取出口IP地址并检查位置
get_exit_ip() {
    echo_info "检查出口IP地址..."
    
    local ip_info
    ip_info=$(curl -s "http://ip-api.com/json/?fields=status,message,country,countryCode,regionName,city,query" 2>/dev/null)
    
    if [ -z "$ip_info" ]; then
        echo_warning "无法获取IP信息，可能影响Claude API访问"
        echo -n -e "${YELLOW}是否仍要继续启动？[y/N]: ${NC}"
        read -r continue_choice
        
        if [[ ! "$continue_choice" =~ ^[yY] ]]; then
            echo_info "已取消启动"
            return 1
        fi
        return 0
    fi
    
    local status
    status=$(echo "$ip_info" | sed -n 's/.*"status":"\([^"]*\)".*/\1/p')
    
    if [ "$status" = "fail" ]; then
        echo_warning "IP查询失败，可能影响Claude API访问"
        echo -n -e "${YELLOW}是否仍要继续启动？[y/N]: ${NC}"
        read -r continue_choice
        
        if [[ ! "$continue_choice" =~ ^[yY] ]]; then
            echo_info "已取消启动"
            return 1
        fi
        return 0
    fi
    
    local ip country city country_code
    ip=$(echo "$ip_info" | sed -n 's/.*"query":"\([^"]*\)".*/\1/p')
    country=$(echo "$ip_info" | sed -n 's/.*"country":"\([^"]*\)".*/\1/p')
    city=$(echo "$ip_info" | sed -n 's/.*"city":"\([^"]*\)".*/\1/p')
    country_code=$(echo "$ip_info" | sed -n 's/.*"countryCode":"\([^"]*\)".*/\1/p')
    
    echo_info "出口IP: $ip"
    echo_info "位置: $country, $city"
    
    # 检查是否在美国
    if [ "$country_code" != "US" ]; then
        echo_warning "当前IP位置不在美国，可能无法直接访问Claude API"
        echo -n -e "${YELLOW}是否仍要继续启动？[y/N]: ${NC}"
        read -r continue_choice
        
        if [[ ! "$continue_choice" =~ ^[yY] ]]; then
            echo_info "已取消启动"
            return 1
        fi
    else
        echo_success "IP位置在美国，可以正常访问Claude API"
    fi
    
    return 0
}

# 列出配置文件
list_profiles() {
    # 检查是否有上次使用的配置
    local last_used
    last_used=$(get_active_profile)
    
    if [ -n "$last_used" ] && [ -f "$PROFILES_DIR/$last_used.conf" ]; then
        local display_name
        display_name=$(grep "^NAME=" "$PROFILES_DIR/$last_used.conf" 2>/dev/null | cut -d'=' -f2- | tr -d '"' || echo "$last_used")
        
        echo_title "快速启动"
        echo "  上次使用: \"$display_name\""
        echo
        echo "  1 继续使用\"$display_name\""
        echo "  2 选择其他配置"
        echo "  3 退出"
        
        echo -n -e "\n${YELLOW}请选择 [1-3] (默认: 1): ${NC}"
        read -r quick_choice
        
        # 如果用户直接按回车，使用默认选项1
        case "${quick_choice:-1}" in
            1)
                run_claude_with_profile "$last_used"
                return
                ;;
            2)
                # 继续显示完整配置列表
                ;;
            3)
                echo_info "再见！"
                exit 0
                ;;
            *)
                echo_warning "无效选择，显示完整配置列表"
                ;;
        esac
    fi
    
    echo_title "可用配置"
    
    local profiles=()
    local count=1
    
    if [ -d "$PROFILES_DIR" ]; then
        for config_file in "$PROFILES_DIR"/*.conf; do
            if [ -f "$config_file" ]; then
                local name
                name=$(basename "$config_file" .conf)
                profiles+=("$name")
                
                # 读取配置名称
                local display_name=""
                if [ -f "$config_file" ]; then
                    display_name=$(grep "^NAME=" "$config_file" 2>/dev/null | cut -d'=' -f2- | tr -d '"' || echo "")
                fi
                
                # 如果没有NAME字段，使用文件名
                if [ -z "$display_name" ]; then
                    display_name="$name"
                fi
                
                echo "  $count 使用\"$display_name\""
                ((count++))
            fi
        done
    fi
    
    echo "  $count 创建新配置"
    ((count++))
    echo "  $count 退出"
    
    echo -n -e "\n${YELLOW}请选择配置 [1-$count]: ${NC}"
    read -r choice
    
    if [[ "$choice" =~ ^[0-9]+$ ]] && [ "$choice" -ge 1 ] && [ "$choice" -le $count ]; then
        if [ "$choice" -eq $count ]; then
            # 退出
            echo_info "再见！"
            exit 0
        elif [ "$choice" -eq $((count-1)) ]; then
            # 创建新配置
            create_new_profile
        else
            # 选择已有配置
            local selected_profile="${profiles[$((choice-1))]}"
            set_active_profile "$selected_profile"
            handle_existing_profile "$selected_profile"
        fi
    else
        echo_error "无效的选择"
        exit 1
    fi
}

# 处理已有配置
handle_existing_profile() {
    local profile_name="$1"
    local config_file="$PROFILES_DIR/$profile_name.conf"
    
    echo_title "配置: $profile_name"
    
    # 显示配置概要
    show_profile_summary "$config_file"
    
    echo -e "\n${YELLOW}选择操作:${NC}"
    echo "  1 启动 Claude"
    echo "  2 查看/编辑配置"
    echo "  3 删除此配置"
    echo "  4 返回主菜单"
    echo "  5 退出"
    
    echo -n -e "\n${YELLOW}请选择 [1-5] (默认: 1): ${NC}"
    read -r action
    
    case "${action:-1}" in
        1)
            run_claude_with_profile "$profile_name"
            ;;
        2)
            edit_profile "$config_file"
            # 编辑后返回主菜单
            show_main_menu
            ;;
        3)
            delete_profile "$profile_name"
            ;;
        4)
            # 返回主菜单
            show_main_menu
            ;;
        5)
            echo_info "再见！"
            exit 0
            ;;
        *)
            echo_error "无效的选择"
            exit 1
            ;;
    esac
}

# 显示配置概要
show_profile_summary() {
    local config_file="$1"
    
    if [ ! -f "$config_file" ]; then
        echo_warning "配置文件不存在"
        return
    fi
    
    echo_info "配置概要:"
    
    # 安全地读取配置变量
    local auth_token base_url proxy_url
    
    # 使用grep和sed安全解析配置文件
    auth_token=$(grep "^ANTHROPIC_AUTH_TOKEN=" "$config_file" 2>/dev/null | cut -d'=' -f2- | tr -d '"' || echo "")
    base_url=$(grep "^ANTHROPIC_BASE_URL=" "$config_file" 2>/dev/null | cut -d'=' -f2- | tr -d '"' || echo "")
    proxy_url=$(grep "^http_proxy=" "$config_file" 2>/dev/null | cut -d'=' -f2- | tr -d '"' || echo "")
    
    if [ -n "$base_url" ]; then
        echo "  Base URL: $base_url"
    else
        echo "  Base URL: 默认 (api.anthropic.com)"
    fi
    
    if [ -n "$auth_token" ]; then
        local masked_token
        masked_token=$(echo "$auth_token" | sed 's/\(.\{6\}\).*/\1***/')
        echo "  Auth Token: $masked_token"
    else
        echo "  Auth Token: 未设置"
    fi
    
    if [ -n "$proxy_url" ]; then
        echo "  代理: $proxy_url"
    else
        echo "  代理: 未设置"
    fi
}

# 编辑配置
edit_profile() {
    local config_file="$1"
    
    echo_info "使用 vi 编辑配置文件..."
    
    # 备份原文件
    cp "$config_file" "${config_file}.backup"
    
    # 使用vi编辑
    vi "$config_file"
    
    echo_success "配置已更新"
}

# 删除配置
delete_profile() {
    local profile_name="$1"
    local config_file="$PROFILES_DIR/$profile_name.conf"
    
    if [ ! -f "$config_file" ]; then
        echo_error "配置文件不存在"
        show_main_menu
        return
    fi
    
    echo_warning "确认删除配置 '$profile_name'？"
    echo_info "此操作无法撤销！"
    echo -n -e "${YELLOW}请输入 'yes' 确认删除: ${NC}"
    read -r confirm
    
    if [ "$confirm" = "yes" ]; then
        rm "$config_file"
        
        # 如果删除的是当前激活的配置，清除active文件
        local current_active
        current_active=$(get_active_profile)
        if [ "$current_active" = "$profile_name" ]; then
            rm -f "$ACTIVE_FILE"
            echo_info "已清除活动配置记录"
        fi
        
        echo_success "配置 '$profile_name' 已删除"
    else
        echo_info "取消删除"
    fi
    
    # 返回主菜单
    show_main_menu
}

# 创建新配置
create_new_profile() {
    echo_title "创建新配置"
    
    # 1. 输入名称
    echo -n -e "${YELLOW}请输入配置名称: ${NC}"
    read -r profile_name
    
    if [ -z "$profile_name" ]; then
        echo_error "配置名称不能为空"
        exit 1
    fi
    
    # 检查名称是否已存在
    if [ -f "$PROFILES_DIR/$profile_name.conf" ]; then
        echo_error "配置 '$profile_name' 已存在"
        exit 1
    fi
    
    # 2. 输入 ANTHROPIC_BASE_URL
    echo -n -e "${YELLOW}请输入 ANTHROPIC_BASE_URL (留空使用默认): ${NC}"
    read -r base_url
    
    # 3. 输入 ANTHROPIC_AUTH_TOKEN
    echo -n -e "${YELLOW}请输入 ANTHROPIC_AUTH_TOKEN (留空跳过): ${NC}"
    read -s auth_token
    echo
    
    # 4. 输入代理设置
    echo -n -e "${YELLOW}请输入代理地址 (格式: http://host:port，留空不使用代理): ${NC}"
    read -r proxy
    
    # 生成配置文件
    local config_file="$PROFILES_DIR/$profile_name.conf"
    
    cat > "$config_file" << EOF
# Claude Switcher 配置文件
NAME="$profile_name"
ANTHROPIC_AUTH_TOKEN="$auth_token"
EOF

    if [ -n "$base_url" ]; then
        echo "ANTHROPIC_BASE_URL=\"$base_url\"" >> "$config_file"
    fi
    
    if [ -n "$proxy" ]; then
        cat >> "$config_file" << EOF
http_proxy="$proxy"
https_proxy="$proxy"
EOF
    fi
    
    cat >> "$config_file" << EOF

# 其他环境变量可以在此添加
EOF
    
    # 设置文件权限
    chmod 600 "$config_file"
    
    echo_success "配置 '$profile_name' 创建成功"
    
    # 询问是否立即使用
    echo -e "\n${YELLOW}接下来要做什么？${NC}"
    echo "  1 启动此配置"
    echo "  2 返回主菜单"
    echo "  3 退出"
    
    echo -n -e "\n${YELLOW}请选择 [1-3] (默认: 1): ${NC}"
    read -r next_action
    
    case "${next_action:-1}" in
        1)
            set_active_profile "$profile_name"
            run_claude_with_profile "$profile_name"
            ;;
        2)
            show_main_menu
            ;;
        3)
            echo_info "再见！"
            exit 0
            ;;
        *)
            echo_info "返回主菜单"
            show_main_menu
            ;;
    esac
}

# 使用指定配置启动Claude
run_claude_with_profile() {
    local profile_name="$1"
    local config_file="$PROFILES_DIR/$profile_name.conf"
    
    if [ ! -f "$config_file" ]; then
        echo_error "配置文件不存在: $config_file"
        exit 1
    fi
    
    echo_title "启动 Claude - 配置: $profile_name"
    
    # 保存原始环境变量
    local original_token="$ANTHROPIC_AUTH_TOKEN"
    local original_base_url="$ANTHROPIC_BASE_URL"
    local original_http_proxy="$http_proxy"
    local original_https_proxy="$https_proxy"
    
    # 安全加载配置
    local auth_token base_url proxy_url
    auth_token=$(grep "^ANTHROPIC_AUTH_TOKEN=" "$config_file" 2>/dev/null | cut -d'=' -f2- | tr -d '"')
    base_url=$(grep "^ANTHROPIC_BASE_URL=" "$config_file" 2>/dev/null | cut -d'=' -f2- | tr -d '"')
    proxy_url=$(grep "^http_proxy=" "$config_file" 2>/dev/null | cut -d'=' -f2- | tr -d '"')
    
    # 设置变量
    ANTHROPIC_AUTH_TOKEN="$auth_token"
    ANTHROPIC_BASE_URL="$base_url"
    http_proxy="$proxy_url"
    https_proxy="$proxy_url"
    
    # 导出环境变量
    export ANTHROPIC_AUTH_TOKEN
    if [ -n "$ANTHROPIC_BASE_URL" ]; then
        export ANTHROPIC_BASE_URL
    fi
    if [ -n "$http_proxy" ]; then
        export http_proxy
        export https_proxy
    fi
    
    # 检查出口IP（仅在使用默认API地址时）
    if [ -z "$base_url" ]; then
        echo_info "使用默认API地址，检查网络连通性..."
        if ! get_exit_ip; then
            # 恢复原始环境变量
            export ANTHROPIC_AUTH_TOKEN="$original_token"
            export ANTHROPIC_BASE_URL="$original_base_url" 
            export http_proxy="$original_http_proxy"
            export https_proxy="$original_https_proxy"
            return 0
        fi
    else
        echo_info "使用自定义API地址，跳过IP检查: $base_url"
    fi
    
    echo_info "环境变量已设置"
    
    # 检查Claude CLI是否安装
    if ! command -v claude &> /dev/null; then
        echo_error "Claude CLI 未安装"
        echo_info "请访问 https://github.com/anthropics/claude-cli 安装"
        return 1
    fi
    
    echo_success "正在启动 Claude..."
    echo_info "按 Ctrl+C 退出"
    
    # 设置退出时恢复环境的陷阱
    trap 'echo_info "正在恢复环境..."; export ANTHROPIC_AUTH_TOKEN="$original_token"; export ANTHROPIC_BASE_URL="$original_base_url"; export http_proxy="$original_http_proxy"; export https_proxy="$original_https_proxy"; exit 0' INT TERM
    
    # 启动Claude
    claude
    
    # 恢复原始环境变量
    export ANTHROPIC_AUTH_TOKEN="$original_token"
    export ANTHROPIC_BASE_URL="$original_base_url" 
    export http_proxy="$original_http_proxy"
    export https_proxy="$original_https_proxy"
}

# 显示主菜单
show_main_menu() {
    echo_title "Claude Switcher"
    
    # 显示配置列表和选项
    list_profiles
}

# 主函数
main() {
    # 检查依赖
    check_dependencies
    
    # 初始化配置目录
    init_config_dir
    
    # 解析命令行参数
    parse_arguments "$@"
    
    # 显示主菜单
    show_main_menu
}

# 启动程序
main "$@"