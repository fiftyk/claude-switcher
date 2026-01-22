# 贡献指南

感谢您考虑为 Claude Switcher 做出贡献！本指南将帮助您了解如何参与项目。

## 开发环境设置

### 前置要求

- Go 1.23+
- Git

### 安装开发工具

```bash
# 安装 commitizen (用于规范化提交)
pip install commitizen

# 或者使用 npm
npm install -g commitizen cz-conventional-changelog
```

### 本地开发

```bash
# 克隆项目
git clone https://github.com/fiftyk/claude-switcher.git
cd claude-switcher

# 安装依赖
go mod tidy

# 运行测试
go test ./...

# 构建
go build -o claude-switcher .
```

## 提交规范

本项目使用 [Conventional Commits](https://www.conventionalcommits.org/) 规范。

### 使用 commitizen

```bash
# 交互式提交
cz c

# 版本升级预览 (不实际执行)
cz bump --dry-run

# 版本升级
cz bump
```

### 提交类型

| 类型 | 描述 |
|------|------|
| `feat` | 新功能 |
| `fix` | Bug 修复 |
| `refactor` | 重构 |
| `docs` | 文档更新 |
| `style` | 代码格式（不影响功能） |
| `test` | 测试相关 |
| `chore` | 构建过程或辅助工具更改 |

### 示例

```
feat: 添加配置同步到 settings.json 功能

- 使用 --sync 参数同步配置
- 自动保留 settings.json 中的其他配置
- 添加 _claudeSwitcherProfile 标记

Closes #123
```

## 发布流程

### 自动发布 (推荐)

1. **提交更改** 使用 `cz c`
2. **版本升级** 使用 `cz bump`
3. **推送标签** GitHub Actions 自动构建和发布

### 手动发布

```bash
# 打标签
git tag v2.0.0

# 推送标签
git push origin v2.0.0
```

GitHub Actions 将自动：
- 运行测试
- 构建多平台二进制 (macOS/Linux, amd64/arm64)
- 生成 Changelog
- 创建 Release Draft

## 测试

```bash
# 运行所有测试
go test ./...

# 运行特定包
go test ./internal/config/...

# 查看测试覆盖率
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
```

## 代码风格

- 遵循 Go 官方代码风格
- 使用 `gofmt` 格式化代码
- 添加单元测试

```bash
# 格式化
gofmt -w .

# 检查
go vet ./...
```

## 许可证

MIT License
