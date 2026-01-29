package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/fiftyk/claude-switcher/internal/config"
	"github.com/fiftyk/claude-switcher/internal/profile"
	"github.com/fiftyk/claude-switcher/internal/settings"
)

// ProfileInterface 定义了 profile 的接口
type ProfileInterface interface {
	GetEnvVars() map[string]string
}

// Profile 是对 profile.Profile 的包装
type Profile struct {
	*profile.Profile
}

func (p *Profile) GetEnvVars() map[string]string {
	return p.EnvVars
}

// BuildEnvVarsFromProfile 从 profile 构建环境变量 map
func BuildEnvVarsFromProfile(p *profile.Profile) map[string]string {
	envVars := make(map[string]string)

	if p.AuthToken != "" {
		envVars["ANTHROPIC_AUTH_TOKEN"] = p.AuthToken
	}
	if p.BaseURL != "" {
		envVars["ANTHROPIC_BASE_URL"] = p.BaseURL
	}
	if p.HTTPProxy != "" {
		envVars["http_proxy"] = p.HTTPProxy
		envVars["https_proxy"] = p.HTTPProxy
	}
	if p.Model != "" {
		envVars["ANTHROPIC_MODEL"] = p.Model
	}

	// 添加自定义环境变量
	for k, v := range p.EnvVars {
		envVars[k] = v
	}

	return envVars
}

// GetExitIP 获取出口 IP
func GetExitIP() (string, error) {
	cmd := exec.Command("curl", "-s", "http://ip-api.com/json/?fields=query")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("curl failed")
	}

	// 简单解析 JSON 获取 IP
	ip := string(output)
	// ip 格式可能是 {"query":"1.2.3.4"}
	return ip, nil
}

// PrintHelp 打印帮助信息
func PrintHelp(w *os.File) error {
	helpText := `Claude Switcher - 使用帮助

用法:
  claude-switcher                    启动交互式配置选择
  claude-switcher <配置名称> [-- <参数...>]  使用指定配置启动，可透传参数
  claude-switcher <配置名称> --sync         切换配置并同步到 settings.json
  claude-switcher --config <名称> [-- <参数...>] 使用指定配置启动，可透传参数
  claude-switcher --list             列出所有可用配置
  claude-switcher --test <名称>      测试配置有效性
  claude-switcher --rename <旧> <新> 重命名配置
  claude-switcher --copy <源> <目标>  复制配置
  claude-switcher --help             显示此帮助信息

说明:
  • 配置文件位于: ~/.claude-switcher/profiles/
  • 无参数运行时进入交互式菜单
  • 使用 --sync 参数可将配置同步到 ~/.claude/settings.json
`
	_, err := fmt.Fprint(w, helpText)
	return err
}

// SyncToSettings 将 profile 同步到 settings.json
func SyncToSettings(profileName string, p *profile.Profile) error {
	settingsPath := GetSettingsFilePath()

	// 构建环境变量
	envVars := make(map[string]string)
	if p.AuthToken != "" {
		envVars["ANTHROPIC_AUTH_TOKEN"] = p.AuthToken
	}
	if p.BaseURL != "" {
		envVars["ANTHROPIC_BASE_URL"] = p.BaseURL
	}
	if p.HTTPProxy != "" {
		envVars["http_proxy"] = p.HTTPProxy
		envVars["https_proxy"] = p.HTTPProxy
	}
	for k, v := range p.EnvVars {
		envVars[k] = v
	}

	return settings.SyncProfileToSettings(settingsPath, profileName, envVars)
}

// SetActiveProfile 设置活动配置
func SetActiveProfile(name string) error {
	activeFile := config.GetActiveFile()
	return os.WriteFile(activeFile, []byte(name), 0600)
}

// GetActiveProfile 获取活动配置
func GetActiveProfile() (string, error) {
	activeFile := config.GetActiveFile()
	data, err := os.ReadFile(activeFile)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	return string(data), nil
}

// RunClaude 运行 claude CLI，使用 profile 中的环境变量
func RunClaude(p *profile.Profile, args ...string) error {
	cmd := exec.Command("claude", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// 构建环境变量，profile 配置优先于现有环境
	env := os.Environ()
	profileEnv := BuildEnvVarsFromProfile(p)
	for k, v := range profileEnv {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}
	cmd.Env = env

	return cmd.Run()
}
