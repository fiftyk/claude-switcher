# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.io/).

---

## [v1.1.8] - 2026-01-22

### Fixes
- goreleaser: 使用非草稿模式发布，便于一键安装脚本下载
- goreleaser: 移除已弃用的 -s 编译标志
- goreleaser: 更新 v2 配置格式（移除 deprecated 字段）
- goreleaser: 更新配置文件版本为 version 2
- goreleaser: 修复下载链接使用具体版本号 v2.13.3
- goreleaser: 移除无效的 --skip-validate 参数
- GitHub Actions: 禁用 Go 缓存避免 go.sum 缺失警告

### Docs
- 添加 CLAUDE.md 文件，为 Claude Code 提供项目指南

---

## [v1.0.0] - 2026-01-22

> **Major Update**: 重构为 Go 项目，消除外部依赖，提供更可靠的 JSON 处理。

### Features
- **配置同步到 settings.json** (新增 `--sync` 参数)
  - 使用 `claude-switcher <配置名> --sync` 切换并同步配置
  - 自动将环境变量写入 `~/.claude/settings.json`
  - 保留 settings.json 中的其他配置（如 enabledPlugins）
  - 添加 `_claudeSwitcherProfile` 标记当前配置
- **参数透传功能**
  - 支持通过 `--` 分隔符将参数直接传递给 claude CLI
  - 支持多种透传形式

### Refactor
- 使用 Go 1.23 重构，消除了对 jq 的依赖
- 原生 JSON 支持，更可靠的 settings.json 处理
- TDD 开发模式，所有核心功能有单元测试覆盖
- 预编译二进制，跨平台兼容

### Engineering
- GitHub Actions 自动构建和发布
- 支持 macOS (Intel/Apple Silicon) 和 Linux
- 一键安装脚本自动下载预编译二进制
- commitizen + Go Releaser 自动化版本管理

---

## Format

### Types
- `Features` - New features
- `Fixes` - Bug fixes
- `Refactor` - Code refactoring
- `Docs` - Documentation changes
- `Engineering` - Build/CI changes

### Reference
- [Semantic Versioning](https://semver.org/)
- [Keep a Changelog](https://keepachangelog.com/)
