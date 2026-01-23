package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/fiftyk/claude-switcher/internal/profile"
	"github.com/fiftyk/claude-switcher/internal/settings"
	"gopkg.in/yaml.v3"
)

// ImportFormat 导入格式
type ImportFormat string

const (
	ImportFormatJSON ImportFormat = "json"
	ImportFormatYAML ImportFormat = "yaml"
)

// ImportData JSON/YAML 导入数据结构
type ImportData struct {
	Name        string            `json:"name,omitempty" yaml:"name"`
	AuthToken   string            `json:"auth_token,omitempty" yaml:"auth_token"`
	BaseURL     string            `json:"base_url,omitempty" yaml:"base_url"`
	HTTPProxy   string            `json:"http_proxy,omitempty" yaml:"http_proxy"`
	HTTPSProxy  string            `json:"https_proxy,omitempty" yaml:"https_proxy"`
	Model       string            `json:"model,omitempty" yaml:"model"`
	CustomVars  map[string]string `json:"custom_vars,omitempty" yaml:"custom_vars"`
}

// ImportProfileFromJSON 从 JSON 导入配置
func ImportProfileFromJSON(jsonData, profileName string) (*profile.Profile, error) {
	var data ImportData
	if err := json.Unmarshal([]byte(jsonData), &data); err != nil {
		return nil, fmt.Errorf("无效的 JSON 格式: %w", err)
	}

	p := &profile.Profile{
		Name:       profileName,
		AuthToken:  data.AuthToken,
		BaseURL:    data.BaseURL,
		HTTPProxy:  data.HTTPProxy,
		HTTPSProxy: data.HTTPSProxy,
		Model:      data.Model,
		EnvVars:    make(map[string]string),
	}

	// 使用 JSON 中的名称作为显示名
	if data.Name != "" {
		p.Name = data.Name
	}

	// 添加自定义变量
	for k, v := range data.CustomVars {
		p.EnvVars[k] = v
	}

	return p, nil
}

// ImportProfileFromYAML 从 YAML 导入配置
func ImportProfileFromYAML(yamlData, profileName string) (*profile.Profile, error) {
	var data ImportData
	if err := yaml.Unmarshal([]byte(yamlData), &data); err != nil {
		return nil, fmt.Errorf("无效的 YAML 格式: %w", err)
	}

	p := &profile.Profile{
		Name:       profileName,
		AuthToken:  data.AuthToken,
		BaseURL:    data.BaseURL,
		HTTPProxy:  data.HTTPProxy,
		HTTPSProxy: data.HTTPSProxy,
		Model:      data.Model,
		EnvVars:    make(map[string]string),
	}

	// 使用 YAML 中的名称作为显示名
	if data.Name != "" {
		p.Name = data.Name
	}

	// 添加自定义变量
	for k, v := range data.CustomVars {
		p.EnvVars[k] = v
	}

	return p, nil
}

// SaveProfileToFile 保存配置到文件
func SaveProfileToFile(profilesDir string, p *profile.Profile, fileName string) error {
	// 如果没有提供文件名，使用 profile 的显示名
	if fileName == "" {
		fileName = p.Name
	}

	filePath := filepath.Join(profilesDir, fileName+".conf")

	// 检查是否已存在
	if _, err := os.Stat(filePath); err == nil {
		return fmt.Errorf("配置 '%s' 已存在", fileName)
	}

	content := formatProfile(p)
	if err := os.WriteFile(filePath, []byte(content), 0600); err != nil {
		return fmt.Errorf("无法保存配置文件: %w", err)
	}

	return nil
}

// ImportFromSettings 从 settings.json 导入配置
func ImportFromSettings(settingsPath, profileName string) (*profile.Profile, error) {
	// 加载 settings.json
	s, err := settings.LoadSettings(settingsPath)
	if err != nil {
		return nil, fmt.Errorf("无法加载 settings.json: %w", err)
	}

	p := &profile.Profile{
		Name:       profileName,
		AuthToken:  s.Env["ANTHROPIC_AUTH_TOKEN"],
		BaseURL:    s.Env["ANTHROPIC_BASE_URL"],
		HTTPProxy:  s.Env["http_proxy"],
		HTTPSProxy: s.Env["https_proxy"],
		Model:      s.Env["ANTHROPIC_MODEL"],
		EnvVars:    make(map[string]string),
	}

	// 复制其他环境变量
	for k, v := range s.Env {
		switch k {
		case "ANTHROPIC_AUTH_TOKEN", "ANTHROPIC_BASE_URL", "http_proxy", "https_proxy", "ANTHROPIC_MODEL":
			// 已处理
		default:
			p.EnvVars[k] = v
		}
	}

	return p, nil
}

// PrintImportHelp 打印导入帮助信息
func PrintImportHelp() {
	fmt.Println("\n导入格式:")
	fmt.Println("  json   - 从 JSON 格式导入")
	fmt.Println("  yaml   - 从 YAML 格式导入")
	fmt.Println()
	fmt.Println("用法:")
	fmt.Println("  claude-switcher --import <文件> [格式]")
	fmt.Println("  claude-switcher --import-settings [配置名]  从 settings.json 导入")
	fmt.Println()
}
