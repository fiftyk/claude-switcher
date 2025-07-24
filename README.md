# Claude Switcher

简单易用的Claude配置管理工具 - 让你轻松在不同环境下使用Claude

## 🌟 特性

- 🚀 **一键启动** - 运行 `claude-switcher` 即可开始使用
- ⚡ **命令行参数** - 支持 `claude-switcher moonshot` 直接指定配置启动
- 📋 **配置管理** - 创建、重命名、复制、删除Claude配置
- 🔍 **配置验证** - 自动验证URL格式、代理设置等配置有效性
- 🔄 **快速切换** - 自动记住上次使用的配置，按回车快速启动  
- 🌍 **智能IP检查** - 使用默认API时检查出口IP地址，使用自定义API时自动跳过
- 🛡️ **安全增强** - 配置名称安全验证，防止路径遍历攻击
- 🔒 **安全存储** - 配置文件权限保护，安全的环境变量处理
- 🎯 **默认选项** - 所有菜单支持默认选项，按回车选择最常用操作
- 💡 **灵活配置** - 支持空Token配置，便于创建模板和测试
- 🔧 **脚本友好** - 完全支持在自动化脚本中使用

## 📦 安装

### 一键安装
```bash
curl -sSL https://raw.githubusercontent.com/fiftyk/claude-switcher/main/install.sh | bash
```

### 手动安装
```bash
# 下载脚本
curl -L https://raw.githubusercontent.com/fiftyk/claude-switcher/main/claude-switcher.sh -o claude-switcher.sh

# 设置权限并安装
chmod +x claude-switcher.sh
sudo mv claude-switcher.sh /usr/local/bin/claude-switcher
```

## 🚀 使用方法

### 基本用法
```bash
# 交互式启动（原有方式）
claude-switcher

# 直接指定配置启动
claude-switcher moonshot
claude-switcher --config work
claude-switcher -c production

# 配置管理
claude-switcher --list                    # 查看所有可用配置
claude-switcher --test moonshot           # 测试配置有效性
claude-switcher --rename old new          # 重命名配置
claude-switcher --copy source target      # 复制配置

# 显示帮助信息
claude-switcher --help
```

### 操作流程

#### 🚀 快速启动（命令行参数方式）
```bash
# 直接启动指定配置，无交互
claude-switcher moonshot
```

#### 📋 交互式启动（无参数方式）
1. **首次运行** - 会显示配置选择界面
2. **快速启动** - 再次运行时显示上次使用的配置，按回车即可快速启动  
3. **配置管理** - 选择已有配置后可以启动、编辑或删除，按回车默认启动
4. **创建配置** - 按提示依次输入配置信息，所有字段都可留空
5. **一键启动** - 所有菜单都支持默认选项，直接按回车选择最常用操作

### 交互界面示例

#### 快速启动（再次运行时）
```
=== 快速启动 ===
  上次使用: "我的工作配置"

  1 继续使用"我的工作配置"
  2 选择其他配置
  3 退出

请选择 [1-3] (默认: 1):   # 直接按回车启动
```

#### 首次运行/选择其他配置
```
=== 可用配置 ===
  1 使用"anyrouter配置"
  2 使用"代理配置"  
  3 创建新配置
  4 退出

请选择配置 [1-4]:
```

#### 配置操作
```
=== 配置: anyrouter配置 ===
ℹ 配置概要:
  Base URL: https://anyrouter.top
  Auth Token: sk-ant***
  代理: 未设置

选择操作:
  1 启动 Claude
  2 查看/编辑配置
  3 删除此配置
  4 返回主菜单
  5 退出

请选择 [1-5] (默认: 1):   # 直接按回车启动
```

## 🎯 新功能：命令行参数支持

### 使用场景

#### 🔥 脚本自动化
```bash
#!/bin/bash
# 自动化脚本中直接指定配置
claude-switcher production << EOF
请帮我分析今天的日志文件
EOF
```

#### ⚡ 快速切换
```bash
# 无需进入交互菜单，直接切换配置
claude-switcher moonshot
claude-switcher anyrouter
claude-switcher local-proxy
```

#### 📋 配置管理
```bash
# 查看所有可用配置
claude-switcher --list

# 输出示例：
# === 可用配置列表 ===
#   moonshot - Moonshot配置
#   work - 工作环境配置
#   home - 家庭网络配置
```

### 支持的参数格式
- `claude-switcher <配置名>` - 直接指定配置名称
- `claude-switcher --config <配置名>` - 使用长参数格式
- `claude-switcher -c <配置名>` - 使用短参数格式
- `claude-switcher --list` - 列出所有配置
- `claude-switcher --test <配置名>` - 测试配置有效性
- `claude-switcher --rename <旧名称> <新名称>` - 重命名配置
- `claude-switcher --copy <源名称> <目标名称>` - 复制配置
- `claude-switcher --help` - 显示帮助信息

### 配置管理功能

#### 🔍 配置验证
```bash
# 测试配置是否有效
claude-switcher --test moonshot

# 输出示例：
# ℹ 测试配置: moonshot
# ⚠ 警告: 未设置 AUTH_TOKEN
# ✓ 配置验证通过
```

#### 🔄 配置操作
```bash
# 重命名配置
claude-switcher --rename old-name new-name

# 复制配置
claude-switcher --copy source-config backup-config
```

#### 🛡️ 安全增强
- **配置名称验证**: 防止路径遍历攻击，拒绝包含 `../`、`~`、`$` 等危险字符
- **输入格式验证**: 自动验证URL和代理地址格式
- **安全的环境变量处理**: 程序退出时自动恢复原始环境变量

```bash
# 这些配置名称会被拒绝
claude-switcher "../etc/passwd"  # ✗ 路径遍历攻击
claude-switcher "name with space" # ✗ 包含空格
claude-switcher ".hidden"         # ✗ 以点开头
```

### 错误处理
当指定的配置不存在时，会自动显示可用配置列表：
```bash
$ claude-switcher nonexistent
✗ 配置 'nonexistent' 不存在

ℹ 可用配置:
  moonshot - Moonshot配置
  work - 工作环境配置
```

## 📝 配置文件格式

配置文件使用简单的 Shell 变量格式：

```bash
# Claude Switcher 配置文件
NAME="我的anyrouter配置"
ANTHROPIC_AUTH_TOKEN="sk-ant-xxxx"
ANTHROPIC_BASE_URL="https://anyrouter.top"
http_proxy="http://127.0.0.1:7890"
https_proxy="http://127.0.0.1:7890"
```

## 🔧 创建配置流程

运行程序后选择"创建新配置"，然后按提示输入：

1. **配置名称** - 给配置起个名字，如 "work", "home"
2. **ANTHROPIC_BASE_URL** - API端点，留空使用默认
3. **ANTHROPIC_AUTH_TOKEN** - 你的Claude认证Token，可留空后续编辑
4. **代理设置** - 代理地址，留空不使用代理

**说明：所有字段都可以留空，方便创建配置模板或占位配置。**

## 🌍 常用配置示例

### anyrouter配置
- **Base URL**: `https://anyrouter.top`
- **Token**: 你的anyrouter token
- **代理**: 留空

### 代理配置  
- **Base URL**: 留空(使用默认)
- **Token**: 你的官方Claude token
- **代理**: `http://127.0.0.1:7890`

### 直连配置
- **Base URL**: 留空(使用默认)  
- **Token**: 你的官方Claude token
- **代理**: 留空

## 📁 文件位置

- 配置目录: `~/.claude-switcher/`
- 配置文件: `~/.claude-switcher/profiles/*.conf`

## 🔒 安全说明

- 配置文件权限设置为 600 (仅用户可读写)
- Token 在界面中显示时自动掩码
- 启动前保存原环境变量，退出后自动恢复

## 📋 系统要求

- macOS 或 Linux
- bash, curl 
- Claude CLI (用于启动Claude)

## 🔄 从旧版本迁移

如果你之前使用的是 `claude-proxy-checker`，安装脚本会自动检测并提示删除：

```bash
# 一键安装 - 会自动检测旧版本
curl -sSL https://raw.githubusercontent.com/fiftyk/claude-switcher/main/install.sh | bash
```

安装时会提示：
- 是否删除旧的 `claude-proxy-checker` 脚本
- 是否删除旧的配置文件 `~/.claude_proxy_config`

**手动迁移配置：**
如果保留了旧配置文件，可以手动创建对应的新配置：

```bash
# 查看旧配置
cat ~/.claude_proxy_config

# 创建对应的新配置
claude-switcher  # 选择"创建新配置"
```

## 🛠️ 故障排除

### Claude CLI 未安装
```bash
# 安装Claude CLI
# 访问: https://github.com/anthropics/claude-cli
```

### IP检测问题
- **默认API地址**: 如果IP检测失败或显示非美国位置，程序会提示用户选择是否继续启动Claude
- **自定义API地址**: 设置了 `ANTHROPIC_BASE_URL` 时会自动跳过IP检查，提升启动速度

### 配置文件问题
配置文件位于 `~/.claude-switcher/profiles/`，可以手动编辑或删除

### 重新安装
```bash
# 删除旧版本
sudo rm -f /usr/local/bin/claude-switcher

# 重新安装
curl -sSL https://raw.githubusercontent.com/fiftyk/claude-switcher/main/install.sh | bash
```

## 📄 许可证

MIT License

---

**简单、直观、好用！** 🎯