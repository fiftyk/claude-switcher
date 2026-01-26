# CLAUDE.md - Claude Switcher 项目指南

## 项目概述

Claude Switcher 是一个 Go 语言开发的 CLI 工具，用于管理和切换 Claude Code 的配置。支持：
- 多配置管理（存储在 `~/.claude-switcher/profiles/`）
- 配置同步到 `~/.claude/settings.json`
- 参数透传给 claude CLI

## 常用命令

```bash
# 构建项目
go build -o claude-switcher .

# 运行所有测试
go test ./... -v

# 运行特定包的测试
go test ./internal/config/... -v

# 生成测试覆盖率
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out

# 代码格式化
gofmt -w .

# 代码检查
go vet ./...

# 一键安装
curl -fsSL https://raw.githubusercontent.com/fiftyk/claude-switcher/main/install.sh | bash
```

## 项目架构

### 包结构

```
internal/
├── config/       # 配置目录管理、验证函数
├── profile/      # Profile 加载/保存、列表、重命名、复制
├── settings/     # settings.json 读写、同步
└── update/       # 自动更新检查和安装
main.go           # CLI 入口、参数解析
```

### 关键类型

**profile.Profile** - 配置结构
```go
type Profile struct {
    Name       string            // 显示名称
    AuthToken  string            // Anthropic API Token
    BaseURL    string            // API 基础 URL
    HTTPProxy  string            // HTTP 代理
    HTTPSProxy string            // HTTPS 代理
    EnvVars    map[string]string // 其他环境变量
}
```

**settings.Settings** - settings.json 结构
```go
type Settings struct {
    Env                   map[string]string `json:"env,omitempty"`
    EnabledPlugins        map[string]bool   `json:"enabledPlugins,omitempty"`
    ClaudeSwitcherProfile string           `json:"_claudeSwitcherProfile,omitempty"` // 标记当前激活的配置
}
```

**update.VersionInfo** - 版本信息
```go
type VersionInfo struct {
    Major int
    Minor int
    Patch int
}
```

**update.CheckConfig** - 自动更新检查配置
```go
type CheckConfig struct {
    Repo      string        // GitHub 仓库
    Interval  time.Duration // 检查间隔
    LastCheck time.Time     // 上次检查时间
    Enabled   bool          // 是否启用
}
```

**update.UpdateResult** - 更新检查结果
```go
type UpdateResult struct {
    HasUpdate    bool       // 是否有新版本
    Latest       VersionInfo // 最新版本
    DownloadURL  string     // 下载链接
    ChangelogURL string     // 更新日志链接
}
```

## 版本管理

项目使用 **commitizen + Go Releaser** 实现自动化版本管理：

```bash
# 交互式提交（自动分析提交类型）
cz c

# 版本升级预览（不实际执行）
cz bump --dry-run

# 版本升级（自动打标签）
cz bump
```

**发布流程**：
1. `cz c` 提交更改
2. `cz bump` 升级版本
3. GitHub Actions 自动构建和发布

**版本格式**: `v{major}.{minor}.{patch}`（如 v1.0.0）

## CLI 使用

```bash
# 基本用法
claude-switcher                    # 交互式配置选择
claude-switcher work               # 使用指定配置启动
claude-switcher work --sync        # 切换并同步到 settings.json
claude-switcher work -- -p "提示词" # 透传参数给 claude CLI

# 管理命令
claude-switcher --list             # 列出所有配置
claude-switcher --test <名称>      # 测试配置有效性
claude-switcher --rename <旧> <新>  # 重命名配置
claude-switcher --copy <源> <目标>  # 复制配置

# 自动更新
claude-switcher --check-update     # 检查是否有新版本
claude-switcher --self-update      # 检查并更新到最新版本
```

## 开发注意事项

### 1. 版本变量注入
在 `goreleaser.yml` 中通过 ldflags 注入：
```yaml
ldflags:
  - -s -w -X main.version={{.Version}} -X main.commit={{.Commit}} -X main.date={{.Date}}
```

### 2. Profile 文件格式
使用 Shell 变量格式，存储在 `~/.claude-switcher/profiles/{name}`：
```bash
NAME="工作配置"
ANTHROPIC_AUTH_TOKEN="sk-..."
ANTHROPIC_BASE_URL="https://api.example.com"
http_proxy="http://127.0.0.1:7890"
```

### 3. settings.json 同步
同步时会保留现有配置，仅更新 `env` 对象，并添加 `_claudeSwitcherProfile` 标记。

### 4. Go 版本要求
项目使用 Go 1.23+（见 CONTRIBUTING.md 和 .github/workflows/release.yml）

## 测试策略

- **TDD 方式**：先写测试，再实现功能
- 测试文件与实现文件同名：`*_test.go` 和 `.go`
- 覆盖配置验证、Profile 读写、settings.json 同步等核心逻辑
