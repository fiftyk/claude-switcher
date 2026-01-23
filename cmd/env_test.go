package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/fiftyk/claude-switcher/internal/profile"
)

func TestPreviewEnvVars(t *testing.T) {
	// 创建临时目录和测试配置
	tmpDir := t.TempDir()
	profileName := "test-profile"
	profileFile := filepath.Join(tmpDir, profileName+".conf")

	content := `NAME="Test Profile"
ANTHROPIC_AUTH_TOKEN="sk-test-token"
ANTHROPIC_BASE_URL="https://api.example.com"
http_proxy="http://127.0.0.1:7890"
ANTHROPIC_MODEL="claude-3-5-sonnet"
CUSTOM_VAR="custom-value"
`
	if err := os.WriteFile(profileFile, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	// 加载配置
	p, err := profile.LoadProfile(tmpDir, profileName)
	if err != nil {
		t.Fatal(err)
	}

	// 预览环境变量
	envVars := PreviewEnvVars(p)

	// 验证包含必要变量
	if envVars["ANTHROPIC_AUTH_TOKEN"] != "sk-test-token" {
		t.Errorf("expected ANTHROPIC_AUTH_TOKEN to be sk-test-token, got %s", envVars["ANTHROPIC_AUTH_TOKEN"])
	}
	if envVars["ANTHROPIC_BASE_URL"] != "https://api.example.com" {
		t.Errorf("expected ANTHROPIC_BASE_URL to be https://api.example.com, got %s", envVars["ANTHROPIC_BASE_URL"])
	}
	if envVars["http_proxy"] != "http://127.0.0.1:7890" {
		t.Errorf("expected http_proxy to be http://127.0.0.1:7890, got %s", envVars["http_proxy"])
	}
	if envVars["https_proxy"] != "http://127.0.0.1:7890" {
		t.Errorf("expected https_proxy to be http://127.0.0.1:7890, got %s", envVars["https_proxy"])
	}
	if envVars["ANTHROPIC_MODEL"] != "claude-3-5-sonnet" {
		t.Errorf("expected ANTHROPIC_MODEL to be claude-3-5-sonnet, got %s", envVars["ANTHROPIC_MODEL"])
	}
	if envVars["CUSTOM_VAR"] != "custom-value" {
		t.Errorf("expected CUSTOM_VAR to be custom-value, got %s", envVars["CUSTOM_VAR"])
	}
}

func TestPreviewEnvVarsEmpty(t *testing.T) {
	// 测试空配置
	p := &profile.Profile{
		Name:       "Empty",
		AuthToken:  "",
		BaseURL:    "",
		HTTPProxy:  "",
		HTTPSProxy: "",
		Model:      "",
		EnvVars:    map[string]string{},
	}

	envVars := PreviewEnvVars(p)

	// 应该返回空 map 或只包含基本变量的 map
	if len(envVars) > 3 {
		t.Errorf("expected at most 3 env vars for empty profile, got %d", len(envVars))
	}
}

func TestFormatEnvVarsForDisplay(t *testing.T) {
	envVars := map[string]string{
		"ANTHROPIC_AUTH_TOKEN": "sk-test-token",
		"ANTHROPIC_BASE_URL":   "https://api.example.com",
		"http_proxy":           "http://127.0.0.1:7890",
		"https_proxy":          "http://127.0.0.1:7890",
	}

	output := FormatEnvVarsForDisplay(envVars)

	// 验证输出格式
	if !contains(output, "ANTHROPIC_AUTH_TOKEN") {
		t.Error("output should contain ANTHROPIC_AUTH_TOKEN")
	}
	if !contains(output, "sk-t") {
		t.Error("output should contain masked token")
	}
	if !contains(output, "https_proxy") {
		t.Error("output should contain https_proxy")
	}
}

func TestMaskEnvValue(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		value    string
		expected string
	}{
		{"token", "ANTHROPIC_AUTH_TOKEN", "sk-test-token", "sk-t*******oken"},
		{"url", "ANTHROPIC_BASE_URL", "https://api.example.com", "https://api.example.com"},
		{"proxy", "http_proxy", "http://127.0.0.1:7890", "http://127.0.0.1:7890"},
		{"model", "ANTHROPIC_MODEL", "claude-3-5-sonnet", "claude-3-5-sonnet"},
		{"custom", "CUSTOM_VAR", "custom-value", "custom-value"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MaskEnvValue(tt.key, tt.value)
			if result != tt.expected {
				t.Errorf("MaskEnvValue(%s, %s) = %s, want %s", tt.key, tt.value, result, tt.expected)
			}
		})
	}
}

func TestGenerateExportCommand(t *testing.T) {
	envVars := map[string]string{
		"ANTHROPIC_AUTH_TOKEN": "sk-test",
		"ANTHROPIC_BASE_URL":   "https://api.example.com",
	}

	output := GenerateExportCommand(envVars)

	// 验证输出格式
	if !contains(output, "export") {
		t.Error("output should contain 'export'")
	}
	if !contains(output, "ANTHROPIC_AUTH_TOKEN") {
		t.Error("output should contain ANTHROPIC_AUTH_TOKEN")
	}
}
