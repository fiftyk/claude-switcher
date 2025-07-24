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

# 安全验证函数
validate_config_name() {
    local name="$1"
    
    # 检查是否为空
    if [ -z "$name" ]; then
        return 1
    fi
    
    # 检查长度限制
    if [ ${#name} -gt 50 ]; then
        echo_error "配置名称过长，请限制在50个字符以内"
        return 1
    fi
    
    # 检查是否包含危险字符
    case "$name" in
        */*|*\\*|*..*|*~*|*\$*)
            echo_error "配置名称不能包含特殊字符: / \\ .. ~ \$"
            return 1
            ;;
        .*|*.)
            echo_error "配置名称不能以点开头或结尾"
            return 1
            ;;
        *[[:space:]]*)
            echo_error "配置名称不能包含空格"
            return 1
            ;;
    esac
    
    return 0
}

# 验证URL格式
validate_url() {
    local url="$1"
    
    if [ -z "$url" ]; then
        return 0  # 空值允许
    fi
    
    # 基本URL格式检查
    if [[ ! "$url" =~ ^https?://[a-zA-Z0-9.-]+([:/][^[:space:]]*)?$ ]]; then
        echo_error "URL格式不正确，请使用 http:// 或 https:// 开头"
        return 1
    fi
    
    return 0
}

# 验证代理格式
validate_proxy() {
    local proxy="$1"
    
    if [ -z "$proxy" ]; then
        return 0  # 空值允许
    fi
    
    # 基本代理格式检查
    if [[ ! "$proxy" =~ ^https?://[a-zA-Z0-9.-]+:[0-9]+$ ]]; then
        echo_error "代理格式不正确，请使用格式: http://host:port"
        return 1
    fi
    
    return 0
}

# 安全地恢复环境变量
restore_env_var() {
    local var_name="$1"
    local original_value="$2"
    
    if [ -z "$original_value" ]; then
        unset "$var_name"
    else
        export "$var_name"="$original_value"
    fi
}

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
                --test)
                    echo_error "缺少配置名称"
                    echo_info "用法: claude-switcher --test <配置名称>"
                    exit 1
                    ;;
                --rename)
                    echo_error "缺少参数"
                    echo_info "用法: claude-switcher --rename <旧名称> <新名称>"
                    exit 1
                    ;;
                --copy)
                    echo_error "缺少参数"
                    echo_info "用法: claude-switcher --copy <源名称> <目标名称>"
                    exit 1
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
                --test)
                    test_config "$2"
                    exit $?
                    ;;
                *)
                    echo_error "未知参数组合: $1 $2"
                    show_help_info
                    exit 1
                    ;;
            esac
            ;;
        3)
            case "$1" in
                --rename)
                    rename_config "$2" "$3"
                    exit $?
                    ;;
                --copy)
                    copy_config "$2" "$3"
                    exit $?
                    ;;
                *)
                    echo_error "未知参数组合: $1 $2 $3"
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
    echo "  claude-switcher --test <名称>      测试配置有效性"
    echo "  claude-switcher --rename <旧> <新> 重命名配置"
    echo "  claude-switcher --copy <源> <目标>  复制配置"
    echo "  claude-switcher --help             显示此帮助信息"
    echo
    echo -e "${YELLOW}示例:${NC}"
    echo "  claude-switcher moonshot           # 直接启动moonshot配置"
    echo "  claude-switcher --config work      # 启动work配置"
    echo "  claude-switcher --list             # 查看所有配置"
    echo "  claude-switcher --test moonshot    # 测试配置是否有效"
    echo "  claude-switcher --rename old new   # 重命名配置"
    echo "  claude-switcher --copy work home   # 复制配置"
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
    
    # 验证配置名称安全性
    if ! validate_config_name "$config_name"; then
        echo_error "配置名称格式不正确"
        exit 1
    fi
    
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

# 测试配置有效性
test_config() {
    local config_name="$1"
    local config_file="$PROFILES_DIR/$config_name.conf"
    
    if [ ! -f "$config_file" ]; then
        echo_error "配置文件不存在: $config_name"
        return 1
    fi
    
    echo_info "测试配置: $config_name"
    
    # 读取配置
    local auth_token base_url proxy_url
    auth_token=$(grep "^ANTHROPIC_AUTH_TOKEN=" "$config_file" 2>/dev/null | cut -d'=' -f2- | tr -d '"')
    base_url=$(grep "^ANTHROPIC_BASE_URL=" "$config_file" 2>/dev/null | cut -d'=' -f2- | tr -d '"')
    proxy_url=$(grep "^http_proxy=" "$config_file" 2>/dev/null | cut -d'=' -f2- | tr -d '"')
    
    # 验证配置
    local has_error=false
    
    if [ -z "$auth_token" ]; then
        echo_warning "警告: 未设置 AUTH_TOKEN"
    fi
    
    if [ -n "$base_url" ] && ! validate_url "$base_url"; then
        echo_error "BASE_URL 格式错误: $base_url"
        has_error=true
    fi
    
    if [ -n "$proxy_url" ] && ! validate_proxy "$proxy_url"; then
        echo_error "代理格式错误: $proxy_url"
        has_error=true
    fi
    
    if [ "$has_error" = true ]; then
        echo_error "配置验证失败"
        return 1
    else
        echo_success "配置验证通过"
        return 0
    fi
}

# 重命名配置
rename_config() {
    local old_name="$1"
    local new_name="$2"
    
    if ! validate_config_name "$old_name" || ! validate_config_name "$new_name"; then
        echo_error "配置名称格式不正确"
        return 1
    fi
    
    local old_file="$PROFILES_DIR/$old_name.conf"
    local new_file="$PROFILES_DIR/$new_name.conf"
    
    if [ ! -f "$old_file" ]; then
        echo_error "源配置不存在: $old_name"
        return 1
    fi
    
    if [ -f "$new_file" ]; then
        echo_error "目标配置已存在: $new_name"
        return 1
    fi
    
    # 复制文件并更新NAME字段
    cp "$old_file" "$new_file"
    if grep -q "^NAME=" "$new_file"; then
        sed -i.bak "s/^NAME=.*/NAME=\"$new_name\"/" "$new_file"
        rm "$new_file.bak"
    else
        echo "NAME=\"$new_name\"" >> "$new_file"
    fi
    
    # 删除原文件
    rm "$old_file"
    
    # 更新活动配置
    local current_active
    current_active=$(get_active_profile)
    if [ "$current_active" = "$old_name" ]; then
        set_active_profile "$new_name"
    fi
    
    echo_success "配置已重命名: $old_name -> $new_name"
}

# 复制配置
copy_config() {
    local source_name="$1"
    local target_name="$2"
    
    if ! validate_config_name "$source_name" || ! validate_config_name "$target_name"; then
        echo_error "配置名称格式不正确"
        return 1
    fi
    
    local source_file="$PROFILES_DIR/$source_name.conf"
    local target_file="$PROFILES_DIR/$target_name.conf"
    
    if [ ! -f "$source_file" ]; then
        echo_error "源配置不存在: $source_name"
        return 1
    fi
    
    if [ -f "$target_file" ]; then
        echo_error "目标配置已存在: $target_name"
        return 1
    fi
    
    # 复制文件并更新NAME字段
    cp "$source_file" "$target_file"
    if grep -q "^NAME=" "$target_file"; then
        sed -i.bak "s/^NAME=.*/NAME=\"$target_name\"/" "$target_file"
        rm "$target_file.bak"
    else
        echo "NAME=\"$target_name\"" >> "$target_file"
    fi
    
    echo_success "配置已复制: $source_name -> $target_name"
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
    
    local profile_name base_url auth_token proxy
    
    # 1. 输入和验证配置名称
    while true; do
        echo -n -e "${YELLOW}请输入配置名称: ${NC}"
        read -r profile_name
        
        if validate_config_name "$profile_name"; then
            # 检查名称是否已存在
            if [ -f "$PROFILES_DIR/$profile_name.conf" ]; then
                echo_error "配置 '$profile_name' 已存在"
                echo -n -e "${YELLOW}是否重新输入？[y/N]: ${NC}"
                read -r retry
                if [[ ! "$retry" =~ ^[yY] ]]; then
                    echo_info "返回主菜单"
                    show_main_menu
                    return
                fi
                continue
            fi
            break
        else
            echo -n -e "${YELLOW}是否重新输入？[y/N]: ${NC}"
            read -r retry
            if [[ ! "$retry" =~ ^[yY] ]]; then
                echo_info "返回主菜单"
                show_main_menu
                return
            fi
        fi
    done
    
    # 2. 输入和验证 ANTHROPIC_BASE_URL
    while true; do
        echo -n -e "${YELLOW}请输入 ANTHROPIC_BASE_URL (留空使用默认): ${NC}"
        read -r base_url
        
        if validate_url "$base_url"; then
            break
        else
            echo -n -e "${YELLOW}是否重新输入？[y/N]: ${NC}"
            read -r retry
            if [[ ! "$retry" =~ ^[yY] ]]; then
                base_url=""
                break
            fi
        fi
    done
    
    # 3. 输入 ANTHROPIC_AUTH_TOKEN
    echo -n -e "${YELLOW}请输入 ANTHROPIC_AUTH_TOKEN (留空跳过): ${NC}"
    read -s auth_token
    echo
    
    # 4. 输入和验证代理设置
    while true; do
        echo -n -e "${YELLOW}请输入代理地址 (格式: http://host:port，留空不使用代理): ${NC}"
        read -r proxy
        
        if validate_proxy "$proxy"; then
            break
        else
            echo -n -e "${YELLOW}是否重新输入？[y/N]: ${NC}"
            read -r retry
            if [[ ! "$retry" =~ ^[yY] ]]; then
                proxy=""
                break
            fi
        fi
    done
    
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
            restore_env_var ANTHROPIC_AUTH_TOKEN "$original_token"
            restore_env_var ANTHROPIC_BASE_URL "$original_base_url" 
            restore_env_var http_proxy "$original_http_proxy"
            restore_env_var https_proxy "$original_https_proxy"
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
    trap 'echo_info "正在恢复环境..."; restore_env_var ANTHROPIC_AUTH_TOKEN "$original_token"; restore_env_var ANTHROPIC_BASE_URL "$original_base_url"; restore_env_var http_proxy "$original_http_proxy"; restore_env_var https_proxy "$original_https_proxy"; exit 0' INT TERM
    
    # 启动Claude
    claude
    
    # 恢复原始环境变量
    restore_env_var ANTHROPIC_AUTH_TOKEN "$original_token"
    restore_env_var ANTHROPIC_BASE_URL "$original_base_url" 
    restore_env_var http_proxy "$original_http_proxy"
    restore_env_var https_proxy "$original_https_proxy"
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