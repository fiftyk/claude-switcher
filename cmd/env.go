package cmd

import (
	"fmt"
	"strings"

	"github.com/fiftyk/claude-switcher/internal/profile"
)

// PreviewEnvVars 预览配置将设置的环境变量
func PreviewEnvVars(p *profile.Profile) map[string]string {
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

// FormatEnvVarsForDisplay 格式化环境变量用于显示（遮蔽敏感信息）
func FormatEnvVarsForDisplay(envVars map[string]string) string {
	var sb strings.Builder

	sb.WriteString("\n环境变量:\n")
	sb.WriteString(strings.Repeat("-", 50) + "\n")

	for k, v := range envVars {
		masked := MaskEnvValue(k, v)
		sb.WriteString(fmt.Sprintf("  %-25s = %s\n", k, masked))
	}

	sb.WriteString(strings.Repeat("-", 50) + "\n")
	sb.WriteString("\n使用 'eval \"$(claude-switcher <配置名> --env)\"' 设置环境变量\n")

	return sb.String()
}

// MaskEnvValue 遮蔽敏感的环境变量值
func MaskEnvValue(key, value string) string {
	// 只遮蔽 Auth Token
	if strings.Contains(strings.ToLower(key), "auth_token") ||
	   strings.Contains(strings.ToLower(key), "api_key") ||
	   strings.Contains(strings.ToLower(key), "token") {
		return maskToken(value)
	}
	return value
}

// GenerateExportCommand 生成 export 命令
func GenerateExportCommand(envVars map[string]string) string {
	var sb strings.Builder

	sb.WriteString("# 导出环境变量\n")
	for k, v := range envVars {
		sb.WriteString(fmt.Sprintf("export %s=\"%s\"\n", k, v))
	}

	return sb.String()
}

// PrintEnvPreview 打印环境变量预览
func PrintEnvPreview(p *profile.Profile) {
	envVars := PreviewEnvVars(p)

	fmt.Printf("\n=== 环境变量预览: %s ===\n", p.Name)
	fmt.Println(FormatEnvVarsForDisplay(envVars))
}

// EnvActionType 环境变量操作类型
type EnvActionType int

const (
	EnvActionPreview EnvActionType = iota
	EnvActionExport
	EnvActionEval
)

// ProcessEnvAction 处理环境变量相关操作
func ProcessEnvAction(profilesDir, profileName string, action EnvActionType) error {
	p, err := profile.LoadProfile(profilesDir, profileName)
	if err != nil {
		return fmt.Errorf("无法加载配置: %w", err)
	}

	envVars := PreviewEnvVars(p)

	switch action {
	case EnvActionPreview:
		PrintEnvPreview(p)
	case EnvActionExport:
		fmt.Print(GenerateExportCommand(envVars))
	case EnvActionEval:
		fmt.Printf("eval \"$(claude-switcher %s --env)\"  # 当前不支持\n", profileName)
	}

	return nil
}
