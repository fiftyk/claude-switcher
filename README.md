# Claude Switcher

简单易用的 Claude 配置管理工具 - 让你轻松在不同环境下使用 Claude

## 特性

- **一键启动** - 运行 `claude-switcher` 即可开始使用
- **命令行参数** - 支持 `claude-switcher moonshot` 直接指定配置启动
- **配置同步** - 使用 `--sync` 将配置同步到 `~/.claude/settings.json`
- **配置管理** - 创建、重命名、复制、删除 Claude 配置
- **配置验证** - 自动验证 URL 格式、代理设置等配置有效性
- **快速切换** - 自动记住上次使用的配置，按回车快速启动
- **参数透传** - 支持将参数直接传递给 Claude CLI
- **安全存储** - 配置文件权限保护，安全的环境变量处理

## 安装

### 一键安装 (推荐)

```bash
curl -fsSL https://raw.githubusercontent.com/fiftyk/claude-switcher/main/install.sh | bash
```

### 手动安装

```bash
# macOS (Apple Silicon)
curl -fsSL https://github.com/fiftyk/claude-switcher/releases/download/v1.0.0/claude-switcher-darwin-arm64 -o /usr/local/bin/claude-switcher
chmod +x /usr/local/bin/claude-switcher

# macOS (Intel)
curl -fsSL https://github.com/fiftyk/claude-switcher/releases/download/v1.0.0/claude-switcher-darwin-amd64 -o /usr/local/bin/claude-switcher
chmod +x /usr/local/bin/claude-switcher

# Linux
curl -fsSL https://github.com/fiftyk/claude-switcher/releases/download/v1.0.0/claude-switcher-linux-amd64 -o /usr/local/bin/claude-switcher
chmod +x /usr/local/bin/claude-switcher
```

## 使用方法

### 基本命令

```bash
# 交互式启动
claude-switcher

# 直接指定配置启动
claude-switcher moonshot
claude-switcher --config work

# 同步配置到 settings.json (新功能!)
claude-switcher moonshot --sync

# 列出所有配置
claude-switcher --list

# 测试配置有效性
claude-switcher --test moonshot

# 重命名配置
claude-switcher --rename old new

# 复制配置
claude-switcher --copy source target

# 显示帮助
claude-switcher --help
```

### 参数透传

支持通过 `--` 分隔符将参数直接传递给 Claude CLI：

```bash
# 透传帮助参数
claude-switcher moonshot -- --help

# 透传模型参数
claude-switcher work -- --model claude-3-5-sonnet-20240620

# 透传多个参数
claude-switcher anyrouter -- --temperature 0.7 --max-tokens 1000

# 进入选择菜单后透传参数
claude-switcher -- --version
```

### 新功能：同步到 settings.json

使用 `--sync` 参数可以将当前配置同步到 `~/.claude/settings.json`，使配置持久化：

```bash
# 切换配置并同步到 settings.json
claude-switcher work --sync

# 验证同步结果
cat ~/.claude/settings.json
# 输出示例：
# {
#   "env": {
#     "ANTHROPIC_AUTH_TOKEN": "sk-...",
#     "ANTHROPIC_BASE_URL": "https://api.example.com"
#   },
#   "_claudeSwitcherProfile": "work"
# }
```

## 配置文件

配置文件位于 `~/.claude-switcher/profiles/`，使用简单的变量格式：

```bash
# Claude Switcher 配置文件
NAME="我的配置"
ANTHROPIC_AUTH_TOKEN="sk-ant-xxxx"
ANTHROPIC_BASE_URL="https://api.example.com"
http_proxy="http://127.0.0.1:7890"
https_proxy="http://127.0.0.1:7890"

# 支持任意环境变量
ANTHROPIC_MODEL="claude-3-5-sonnet-20240620"
ANTHROPIC_DEFAULT_HAIKU_MODEL="claude-3-haiku-20240307"
```

## 系统要求

- macOS (Intel/Apple Silicon) 或 Linux
- Claude CLI (用于启动 Claude)

## 从源码构建

```bash
# 克隆项目
git clone https://github.com/fiftyk/claude-switcher.git
cd claude-switcher

# 构建
go build -o claude-switcher .

# 安装
sudo mv claude-switcher /usr/local/bin/
```

## 许可证

MIT License
