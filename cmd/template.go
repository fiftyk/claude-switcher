package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/fiftyk/claude-switcher/internal/profile"
)

// Template 定义配置模板
type Template struct {
	Name        string
	Description string
	Preset      profile.Profile
}

// GetTemplateList 返回所有可用模板
func GetTemplateList() []Template {
	return []Template{
		{
			Name:        "default",
			Description: "默认配置（使用 Anthropic 官方 API）",
			Preset: profile.Profile{
				Name:       "",
				AuthToken:  "",
				BaseURL:    "",
				HTTPProxy:  "",
				HTTPSProxy: "",
				Model:      "",
				EnvVars:    map[string]string{},
			},
		},
		{
			Name:        "openai-compatible",
			Description: "OpenAI 兼容 API（如 OpenRouter、Cloudflare）",
			Preset: profile.Profile{
				Name:       "",
				AuthToken:  "",
				BaseURL:    "https://openrouter.ai/api/v1",
				HTTPProxy:  "",
				HTTPSProxy: "",
				Model:      "",
				EnvVars:    map[string]string{},
			},
		},
		{
			Name:        "proxy",
			Description: "代理配置示例",
			Preset: profile.Profile{
				Name:       "",
				AuthToken:  "",
				BaseURL:    "",
				HTTPProxy:  "http://127.0.0.1:7890",
				HTTPSProxy: "http://127.0.0.1:7890",
				Model:      "",
				EnvVars:    map[string]string{},
			},
		},
		{
			Name:        "claude-code",
			Description: "Claude Code 专用配置",
			Preset: profile.Profile{
				Name:       "",
				AuthToken:  "",
				BaseURL:    "",
				HTTPProxy:  "",
				HTTPSProxy: "",
				Model:      "",
				EnvVars:    map[string]string{
					"CLAUDE_CODE_CFG": "default",
				},
			},
		},
		{
			Name:        "custom-model",
			Description: "自定义模型配置",
			Preset: profile.Profile{
				Name:       "",
				AuthToken:  "",
				BaseURL:    "",
				HTTPProxy:  "",
				HTTPSProxy: "",
				Model:      "claude-3-5-sonnet-20241022",
				EnvVars:    map[string]string{},
			},
		},
	}
}

// GetTemplateByName 根据名称获取模板
func GetTemplateByName(name string) *Template {
	for i := range GetTemplateList() {
		if GetTemplateList()[i].Name == name {
			return &GetTemplateList()[i]
		}
	}
	return nil
}

// ApplyTemplate 应用模板创建配置
func ApplyTemplate(templateName, profileName string) *profile.Profile {
	tmpl := GetTemplateByName(templateName)
	if tmpl == nil {
		return nil
	}

	p := tmpl.Preset
	p.Name = profileName
	return &p
}

// PrintTemplates 打印所有可用模板
func PrintTemplates() {
	templates := GetTemplateList()
	fmt.Println("\n可用模板:")
	fmt.Println(strings.Repeat("-", 50))
	for i, tmpl := range templates {
		fmt.Printf("  %d. %s\n", i+1, tmpl.Name)
		fmt.Printf("     %s\n", tmpl.Description)
	}
	fmt.Println(strings.Repeat("-", 50))
	fmt.Println()
}

// SelectTemplateInteractive 交互式选择模板
func SelectTemplateInteractive() (*Template, error) {
	templates := GetTemplateList()
	if len(templates) == 0 {
		return nil, fmt.Errorf("没有可用模板")
	}

	PrintTemplates()

	fmt.Print("请选择模板编号: ")
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	var idx int
	if _, err := fmt.Sscanf(input, "%d", &idx); err != nil || idx < 1 || idx > len(templates) {
		return nil, fmt.Errorf("无效选择")
	}

	return &templates[idx-1], nil
}
