package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fiftyk/claude-switcher/internal/profile"
)

// ReaderProvider 定义了读取用户输入的接口
type ReaderProvider interface {
	ReadString(delim byte) (string, error)
}

// StdioReader 使用标准输入
type StdioReader struct{}

func (r StdioReader) ReadString(delim byte) (string, error) {
	return bufio.NewReader(os.Stdin).ReadString(delim)
}

// EditProfileInteractive 交互式编辑配置
func EditProfileInteractive(profilesDir, name string) error {
	return EditProfileInteractiveWithReader(profilesDir, name, StdioReader{})
}

// EditProfileInteractiveWithReader 交互式编辑配置（可注入 Reader 进行测试）
func EditProfileInteractiveWithReader(profilesDir, name string, reader ReaderProvider) error {
	// 加载现有配置
	p, err := profile.LoadProfile(profilesDir, name)
	if err != nil {
		return err
	}

	fmt.Printf("\n=== 编辑配置: %s ===\n", name)
	fmt.Println("（直接回车保持当前值，输入新值覆盖）")
	fmt.Println()

	// 显示当前值并提示修改
	displayAndPrompt := func(label, currentValue string) string {
		fmt.Printf("%s [%s]: ", label, maskValue(currentValue))
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		if input == "" {
			return currentValue
		}
		return input
	}

	// 显示名称
	fmt.Printf("显示名称 [%s]: ", p.Name)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	if input != "" {
		p.Name = input
	}

	// Auth Token
	p.AuthToken = displayAndPrompt("ANTHROPIC_AUTH_TOKEN", p.AuthToken)

	// Base URL
	p.BaseURL = displayAndPrompt("ANTHROPIC_BASE_URL", p.BaseURL)

	// HTTP Proxy
	p.HTTPProxy = displayAndPrompt("HTTP Proxy", p.HTTPProxy)
	p.HTTPSProxy = p.HTTPProxy

	// Model
	p.Model = displayAndPrompt("ANTHROPIC_MODEL", p.Model)

	// 保存配置
	filePath := filepath.Join(profilesDir, name+".conf")
	content := formatProfile(p)
	if err := os.WriteFile(filePath, []byte(content), 0600); err != nil {
		return err
	}

	fmt.Printf("\n✓ 配置 '%s' 已保存\n", name)
	return nil
}

// maskValue 遮蔽敏感值
func maskValue(value string) string {
	if len(value) <= 4 {
		return "****"
	}
	return value[:2] + strings.Repeat("*", len(value)-4) + value[len(value)-2:]
}
