package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fiftyk/claude-switcher/internal/profile"
)

// ErrQuit 表示用户选择退出
var ErrQuit = errors.New("user quit")

// MenuHandler 处理菜单操作的接口
type MenuHandler interface {
	ShowMenu(profilesDir string) (MenuAction, string, error)
	RunClaude(args ...string) error
	SetActiveProfile(name string) error
}

// SettingsSyncer 同步配置到 settings.json 的接口
type SettingsSyncer interface {
	SyncToSettings(profileName string, p *profile.Profile) error
}

// DefaultMenuHandler 默认实现
type DefaultMenuHandler struct{}

func (h DefaultMenuHandler) ShowMenu(profilesDir string) (MenuAction, string, error) {
	return ShowMenu(profilesDir)
}

func (h DefaultMenuHandler) RunClaude(args ...string) error {
	return RunClaude(args...)
}

func (h DefaultMenuHandler) SetActiveProfile(name string) error {
	return SetActiveProfile(name)
}

// DefaultSettingsSyncer 默认的 SettingsSyncer 实现
type DefaultSettingsSyncer struct{}

func (s DefaultSettingsSyncer) SyncToSettings(profileName string, p *profile.Profile) error {
	return SyncToSettings(profileName, p)
}

// HandleMenuAction 处理菜单返回的操作
func HandleMenuAction(profilesDir string, action MenuAction, name string, handler MenuHandler, syncer SettingsSyncer) error {
	switch action {
	case ActionQuit:
		// 正常退出
		return ErrQuit

	case ActionRun:
		// 运行配置
		p, err := profile.LoadProfile(profilesDir, name)
		if err != nil {
			return fmt.Errorf("加载配置失败: %w", err)
		}

		if err := handler.SetActiveProfile(name); err != nil {
			return fmt.Errorf("设置活动配置失败: %w", err)
		}

		fmt.Printf("使用配置: %s\n", p.Name)
		return handler.RunClaude()

	case ActionRunWithSync:
		// 同步并运行配置
		p, err := profile.LoadProfile(profilesDir, name)
		if err != nil {
			return fmt.Errorf("加载配置失败: %w", err)
		}

		if err := syncer.SyncToSettings(name, p); err != nil {
			return fmt.Errorf("同步到 settings.json 失败: %w", err)
		}
		fmt.Println("✓ 已同步到 settings.json")

		if err := handler.SetActiveProfile(name); err != nil {
			return fmt.Errorf("设置活动配置失败: %w", err)
		}

		fmt.Printf("使用配置: %s\n", p.Name)
		return handler.RunClaude()

	case ActionCreate:
		// 创建新配置
		return CreateProfileInteractive(profilesDir, name)

	case ActionEdit:
		// 编辑配置
		return EditProfileInteractive(profilesDir, name)

	case ActionDelete:
		// 删除配置
		return DeleteProfileInteractive(profilesDir, name)

	case ActionShowDetails:
		// 显示配置详情
		if err := ShowProfileDetails(profilesDir, name); err != nil {
			return err
		}
		fmt.Print("\n按回车键继续...")
		_, _ = bufio.NewReader(os.Stdin).ReadString('\n')
		return nil

	case ActionExport:
		// 导出配置
		return exportProfile(profilesDir, name)

	case ActionNone:
		// 无操作，继续
		return nil

	default:
		return nil
	}
}

// exportProfile 导出配置（内部函数，便于测试）
func exportProfile(profilesDir, name string) error {
	p, err := profile.LoadProfile(profilesDir, name)
	if err != nil {
		return fmt.Errorf("加载配置失败: %w", err)
	}

	fmt.Println("\n选择导出格式:")
	fmt.Println("  1. JSON")
	fmt.Println("  2. YAML")
	fmt.Println("  3. Shell")

	fmt.Print("请选择: ")
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	var format ExportFormat
	switch input {
	case "1":
		format = FormatJSON
	case "2":
		format = FormatYAML
	case "3":
		format = FormatShell
	default:
		return fmt.Errorf("无效选择")
	}

	data, err := ExportProfile(p, format)
	if err != nil {
		return fmt.Errorf("导出失败: %w", err)
	}

	path, err := SaveExportToFile(data, name, format)
	if err != nil {
		return fmt.Errorf("保存失败: %w", err)
	}
	fmt.Printf("✓ 已导出到: %s\n", path)
	return nil
}

// ExportProfileWithFormat 导出配置到指定路径（供测试使用）
func ExportProfileWithFormat(profilesDir, name string, format ExportFormat, outputPath string) error {
	p, err := profile.LoadProfile(profilesDir, name)
	if err != nil {
		return fmt.Errorf("加载配置失败: %w", err)
	}

	data, err := ExportProfile(p, format)
	if err != nil {
		return fmt.Errorf("导出失败: %w", err)
	}

	return os.WriteFile(outputPath, data, 0644)
}

// ExportProfileToPath 保存导出内容到指定路径
func ExportProfileToPath(data []byte, outputPath string) error {
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}
	return os.WriteFile(outputPath, data, 0644)
}
