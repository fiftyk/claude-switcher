package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/fiftyk/claude-switcher/internal/profile"
)

func TestGetSettingsFilePath(t *testing.T) {
	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, ".claude", "settings.json")

	if got := GetSettingsFilePath(); got != expected {
		t.Errorf("GetSettingsFilePath() = %v, want %v", got, expected)
	}
}

func TestBuildEnvVarsFromProfile(t *testing.T) {
	p := &profile.Profile{
		Name:      "test",
		AuthToken: "sk-test",
		BaseURL:   "https://api.example.com",
		HTTPProxy: "http://proxy:8080",
		Model:     "claude-3-5-sonnet",
		EnvVars:   map[string]string{},
	}

	envVars := BuildEnvVarsFromProfile(p)

	if envVars["ANTHROPIC_AUTH_TOKEN"] != "sk-test" {
		t.Errorf("ANTHROPIC_AUTH_TOKEN = %v, want %v", envVars["ANTHROPIC_AUTH_TOKEN"], "sk-test")
	}
	if envVars["ANTHROPIC_BASE_URL"] != "https://api.example.com" {
		t.Errorf("ANTHROPIC_BASE_URL = %v, want %v", envVars["ANTHROPIC_BASE_URL"], "https://api.example.com")
	}
	if envVars["http_proxy"] != "http://proxy:8080" {
		t.Errorf("http_proxy = %v, want %v", envVars["http_proxy"], "http://proxy:8080")
	}
	if envVars["https_proxy"] != "http://proxy:8080" {
		t.Errorf("https_proxy = %v, want %v", envVars["https_proxy"], "http://proxy:8080")
	}
	if envVars["ANTHROPIC_MODEL"] != "claude-3-5-sonnet" {
		t.Errorf("ANTHROPIC_MODEL = %v, want %v", envVars["ANTHROPIC_MODEL"], "claude-3-5-sonnet")
	}
}

func TestBuildEnvVarsFromProfileEmpty(t *testing.T) {
	p := &profile.Profile{
		Name: "empty",
	}

	envVars := BuildEnvVarsFromProfile(p)

	if len(envVars) != 0 {
		t.Errorf("envVars should be empty, got %v", envVars)
	}
}

func TestGetExitIP(t *testing.T) {
	// 这个测试需要网络连接，简单验证函数执行
	ip, err := GetExitIP()
	// 可能失败（网络问题），但不应该是代码错误
	if err != nil && err.Error() != "curl failed" {
		t.Errorf("GetExitIP() unexpected error: %v", err)
	}
	// 如果成功，IP 应该是有效的格式
	if ip != "" && len(ip) < 7 {
		t.Errorf("GetExitIP() returned invalid IP: %v", ip)
	}
}

type mockProfile struct {
	Name       string
	AuthToken  string
	BaseURL    string
	HTTPProxy  string
	HTTPSProxy string
	Model      string
	EnvVars    map[string]string
}

func (m *mockProfile) GetEnvVars() map[string]string {
	return m.EnvVars
}

// 确保 mockProfile 实现了 ProfileInterface
var _ ProfileInterface = (*mockProfile)(nil)
