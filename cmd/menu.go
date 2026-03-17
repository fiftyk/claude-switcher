package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fiftyk/claude-switcher/internal/config"
	"github.com/fiftyk/claude-switcher/internal/profile"
)

// MenuAction 表示菜单操作类型
type MenuAction int

const (
	ActionNone MenuAction = iota
	ActionRun
	ActionEdit
	ActionDelete
	ActionCreate
	ActionImport
	ActionExport
	ActionShowDetails
	ActionSync
	ActionQuit
)

// MenuItem 表示菜单项
type MenuItem struct {
	Name        string
	DisplayName string
	Profile     *profile.Profile
}

// ShowMenu 显示交互式菜单，返回选择的操作和配置名
func ShowMenu(profilesDir string) (MenuAction, string, error) {
	for {
		profiles, err := loadProfiles(profilesDir)
		if err != nil {
			return ActionNone, "", err
		}

		activeProfile, _ := GetActiveProfile()

		fmt.Println("\n" + strings.Repeat("=", 50))
		fmt.Println("         Claude Switcher - 配置管理")
		fmt.Println(strings.Repeat("=", 50))
		fmt.Println()

		// 快速启动区
		fmt.Println("🚀 快速启动")
		if len(profiles) == 0 {
			fmt.Println("  暂无配置，请先创建")
		} else {
			for i, p := range profiles {
				marker := "  "
				if p.Name == activeProfile {
					marker = "✅"
				}
				fmt.Printf("  %s %d. %s\n", marker, i+1, p.Name)
			}
		}
		fmt.Println()

		// 配置管理区
		fmt.Println("⚙️  配置管理")
		fmt.Println("  n. 创建新配置")
		fmt.Println("  e. 编辑配置")
		fmt.Println("  d. 删除配置")
		fmt.Println()

		// 其他功能区
		fmt.Println("📋 其他功能")
		fmt.Println("  i. 配置详情")
		fmt.Println("  s. 同步到 settings.json")
		fmt.Println("  v. 查看环境变量")
		fmt.Println("  t. 导出配置")
		fmt.Println("  h. 帮助")
		fmt.Println("  q. 退出")
		fmt.Println()

		fmt.Printf("请选择操作 [%d-%d/n/e/d/i/s/v/t/h/q]: ", 1, len(profiles))
		fmt.Print("\033[?25h") // 显示光标

		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		input = strings.ToLower(input)

		// 处理选择
		if input == "q" || input == "quit" || input == "exit" {
			return ActionQuit, "", nil
		}

		if input == "h" || input == "help" {
			PrintHelp(os.Stdout)
			continue
		}

		if input == "n" || input == "new" || input == "create" {
			name, err := promptConfigName(profiles)
			if err != nil {
				fmt.Printf("错误: %v\n", err)
				continue
			}
			return ActionCreate, name, nil
		}

		if input == "e" || input == "edit" {
			name, err := selectProfile(profiles, "编辑")
			if err != nil {
				fmt.Printf("错误: %v\n", err)
				continue
			}
			return ActionEdit, name, nil
		}

		if input == "d" || input == "delete" {
			name, err := selectProfile(profiles, "删除")
			if err != nil {
				fmt.Printf("错误: %v\n", err)
				continue
			}
			return ActionDelete, name, nil
		}

		if input == "i" || input == "info" || input == "details" {
			name, err := selectProfile(profiles, "查看详情")
			if err != nil {
				fmt.Printf("错误: %v\n", err)
				continue
			}
			return ActionShowDetails, name, nil
		}

		if input == "s" || input == "sync" {
			name, err := selectProfile(profiles, "同步到 settings.json")
			if err != nil {
				fmt.Printf("错误: %v\n", err)
				continue
			}
			return ActionSync, name, nil
		}

		if input == "v" || input == "vars" || input == "env" {
			name, err := selectProfile(profiles, "查看环境变量")
			if err != nil {
				fmt.Printf("错误: %v\n", err)
				continue
			}
			return ActionShowDetails, name, nil
		}

		if input == "t" || input == "export" {
			name, err := selectProfile(profiles, "导出")
			if err != nil {
				fmt.Printf("错误: %v\n", err)
				continue
			}
			return ActionExport, name, nil
		}

		// 数字选择 - 运行配置
		var idx int
		if _, err := fmt.Sscanf(input, "%d", &idx); err == nil && idx >= 1 && idx <= len(profiles) {
			name := profiles[idx-1].Name
			return ActionRun, name, nil
		}

		fmt.Println("无效选择，请重试")
	}
}

// loadProfiles 加载所有配置
func loadProfiles(profilesDir string) ([]*profile.Profile, error) {
	names, err := profile.ListProfiles(profilesDir)
	if err != nil {
		return nil, err
	}

	var profiles []*profile.Profile
	for _, name := range names {
		p, err := profile.LoadProfile(profilesDir, name)
		if err != nil {
			continue
		}
		profiles = append(profiles, p)
	}
	return profiles, nil
}

// selectProfile 让用户选择一个配置
func selectProfile(profiles []*profile.Profile, purpose string) (string, error) {
	if len(profiles) == 0 {
		return "", fmt.Errorf("没有可用配置")
	}

	fmt.Printf("\n请选择要%s的配置:\n", purpose)
	for i, p := range profiles {
		fmt.Printf("  %d. %s\n", i+1, p.Name)
	}

	fmt.Print("请输入编号: ")
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	var idx int
	if _, err := fmt.Sscanf(input, "%d", &idx); err != nil || idx < 1 || idx > len(profiles) {
		return "", fmt.Errorf("无效选择")
	}

	return profiles[idx-1].Name, nil
}

// promptConfigName 提示用户输入新配置名称
func promptConfigName(profiles []*profile.Profile) (string, error) {
	existingNames := make(map[string]bool)
	for _, p := range profiles {
		existingNames[p.Name] = true
	}

	fmt.Print("\n请输入新配置名称: ")
	reader := bufio.NewReader(os.Stdin)
	name, _ := reader.ReadString('\n')
	name = strings.TrimSpace(name)

	if name == "" {
		return "", fmt.Errorf("名称不能为空")
	}

	if existingNames[name] {
		return "", fmt.Errorf("配置 '%s' 已存在", name)
	}

	if valid, err := config.ValidateConfigName(name); !valid {
		return "", err
	}

	return name, nil
}

// ShowProfileDetails 显示配置详情
func ShowProfileDetails(profilesDir, name string) error {
	p, err := profile.LoadProfile(profilesDir, name)
	if err != nil {
		return err
	}

	fmt.Printf("\n=== 配置详情: %s ===\n", name)
	fmt.Println()
	fmt.Printf("  显示名称: %s\n", p.Name)
	fmt.Printf("  Auth Token: %s\n", maskToken(p.AuthToken))
	fmt.Printf("  Base URL: %s\n", p.BaseURL)
	fmt.Printf("  HTTP Proxy: %s\n", p.HTTPProxy)
	fmt.Printf("  HTTPS Proxy: %s\n", p.HTTPSProxy)
	fmt.Printf("  Model: %s\n", p.Model)

	if len(p.EnvVars) > 0 {
		fmt.Println()
		fmt.Println("  自定义环境变量:")
		for k, v := range p.EnvVars {
			fmt.Printf("    %s: %s\n", k, v)
		}
	}

	fmt.Println()
	return nil
}

// maskToken 遮蔽 token
func maskToken(token string) string {
	if len(token) <= 8 {
		return "****"
	}
	return token[:4] + "*******" + token[len(token)-4:]
}

// CreateProfileInteractive 交互式创建配置
func CreateProfileInteractive(profilesDir, name string) error {
	fmt.Printf("\n=== 创建配置: %s ===\n", name)
	fmt.Println()

	p := &profile.Profile{
		Name: name,
	}

	reader := bufio.NewReader(os.Stdin)

	// 输入显示名称
	fmt.Print("显示名称 (直接回车使用配置名): ")
	displayName, _ := reader.ReadString('\n')
	displayName = strings.TrimSpace(displayName)
	if displayName != "" {
		p.Name = displayName
	}

	// 输入 Auth Token
	fmt.Print("ANTHROPIC_AUTH_TOKEN (可留空): ")
	token, _ := reader.ReadString('\n')
	token = strings.TrimSpace(token)
	p.AuthToken = token

	// 输入 Base URL
	fmt.Print("ANTHROPIC_BASE_URL (直接回车使用默认值): ")
	baseURL, _ := reader.ReadString('\n')
	baseURL = strings.TrimSpace(baseURL)
	p.BaseURL = baseURL

	// 输入代理
	fmt.Print("HTTP Proxy (直接回车不使用代理): ")
	proxy, _ := reader.ReadString('\n')
	proxy = strings.TrimSpace(proxy)
	p.HTTPProxy = proxy
	p.HTTPSProxy = proxy

	// 输入 Model
	fmt.Print("ANTHROPIC_MODEL (直接回车不使用): ")
	model, _ := reader.ReadString('\n')
	model = strings.TrimSpace(model)
	p.Model = model

	// 保存配置
	filePath := filepath.Join(profilesDir, name+".conf")
	content := formatProfile(p)
	if err := os.WriteFile(filePath, []byte(content), 0600); err != nil {
		return err
	}

	fmt.Printf("\n✓ 配置 '%s' 已创建\n", name)
	return nil
}

// formatProfile 将配置格式化为字符串
func formatProfile(p *profile.Profile) string {
	var sb strings.Builder
	sb.WriteString("# Claude Switcher 配置文件\n")
	sb.WriteString("NAME=\"" + p.Name + "\"\n")

	if p.AuthToken != "" {
		sb.WriteString("ANTHROPIC_AUTH_TOKEN=\"" + p.AuthToken + "\"\n")
	}
	if p.BaseURL != "" {
		sb.WriteString("ANTHROPIC_BASE_URL=\"" + p.BaseURL + "\"\n")
	}
	if p.HTTPProxy != "" {
		sb.WriteString("http_proxy=\"" + p.HTTPProxy + "\"\n")
	}
	if p.HTTPSProxy != "" {
		sb.WriteString("https_proxy=\"" + p.HTTPSProxy + "\"\n")
	}
	if p.Model != "" {
		sb.WriteString("ANTHROPIC_MODEL=\"" + p.Model + "\"\n")
	}
	for k, v := range p.EnvVars {
		sb.WriteString(k + "=\"" + v + "\"\n")
	}

	return sb.String()
}

// DeleteProfileInteractive 交互式删除配置
func DeleteProfileInteractive(profilesDir, name string) error {
	fmt.Printf("\n⚠️  确认删除配置 '%s'？ (输入 y 确认): ", name)
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input != "y" && input != "Y" {
		fmt.Println("已取消")
		return nil
	}

	if err := profile.DeleteProfile(profilesDir, name); err != nil {
		return err
	}

	fmt.Printf("✓ 配置 '%s' 已删除\n", name)
	return nil
}
