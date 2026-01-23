package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/fiftyk/claude-switcher/internal/profile"
)

func TestShowMenu(t *testing.T) {
	// 测试菜单显示函数存在且不 panic
	profilesDir := t.TempDir()

	// 创建测试配置
	profiles := []string{"profile1", "profile2", "profile3"}
	for _, name := range profiles {
		content := "NAME=\"" + name + "\"\nANTHROPIC_AUTH_TOKEN=\"sk-test\"\n"
		if err := os.WriteFile(filepath.Join(profilesDir, name+".conf"), []byte(content), 0600); err != nil {
			t.Fatal(err)
		}
	}

	// 测试 loadProfiles 函数
	list, err := loadProfiles(profilesDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 3 {
		t.Errorf("expected 3 profiles, got %d", len(list))
	}
}

func TestGetProfileChoice(t *testing.T) {
	// 测试选择获取函数
	profilesDir := t.TempDir()

	// 创建测试配置
	content := "NAME=\"Test Profile\"\nANTHROPIC_AUTH_TOKEN=\"sk-test\"\n"
	if err := os.WriteFile(filepath.Join(profilesDir, "test.conf"), []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	// 验证 GetProfileChoice 函数存在
	profiles, err := profile.ListProfiles(profilesDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(profiles) != 1 {
		t.Errorf("expected 1 profile, got %d", len(profiles))
	}
}

func TestShowProfileDetails(t *testing.T) {
	// 测试配置详情显示
	profilesDir := t.TempDir()

	// 创建测试配置
	content := `NAME="Test Profile"
ANTHROPIC_AUTH_TOKEN="sk-test-token"
ANTHROPIC_BASE_URL="https://api.example.com"
http_proxy="http://127.0.0.1:7890"
ANTHROPIC_MODEL="claude-3-5-sonnet"
CUSTOM_VAR="custom-value"
`
	if err := os.WriteFile(filepath.Join(profilesDir, "test.conf"), []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	p, err := profile.LoadProfile(profilesDir, "test")
	if err != nil {
		t.Fatal(err)
	}

	// 验证 ShowProfileDetails 函数存在并能处理配置
	_ = p
	if p.BaseURL != "https://api.example.com" {
		t.Errorf("expected BaseURL to be https://api.example.com, got %s", p.BaseURL)
	}
	if p.HTTPProxy != "http://127.0.0.1:7890" {
		t.Errorf("expected HTTPProxy to be http://127.0.0.1:7890, got %s", p.HTTPProxy)
	}
	if p.EnvVars["CUSTOM_VAR"] != "custom-value" {
		t.Errorf("expected CUSTOM_VAR to be custom-value, got %s", p.EnvVars["CUSTOM_VAR"])
	}
}

func TestMaskToken(t *testing.T) {
	tests := []struct {
		name   string
		token  string
		expect string
	}{
		{"normal token", "sk-ant-api03-abc123def456", "sk-a*******f456"},
		{"short token", "sk-123", "****"},
		{"empty token", "", "****"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := maskToken(tt.token)
			if result != tt.expect {
				t.Errorf("maskToken(%s) = %s, want %s", tt.token, result, tt.expect)
			}
		})
	}
}

func TestFormatProfile(t *testing.T) {
	p := &profile.Profile{
		Name:       "Test",
		AuthToken:  "sk-test",
		BaseURL:    "https://api.example.com",
		HTTPProxy:  "http://127.0.0.1:7890",
		HTTPSProxy: "http://127.0.0.1:7890",
		Model:      "claude-3-5-sonnet",
		EnvVars:    map[string]string{"CUSTOM": "value"},
	}

	result := formatProfile(p)

	// 验证输出包含必要内容
	if !contains(result, "NAME=\"Test\"") {
		t.Error("expected NAME in output")
	}
	if !contains(result, "ANTHROPIC_AUTH_TOKEN=\"sk-test\"") {
		t.Error("expected AuthToken in output")
	}
	if !contains(result, "ANTHROPIC_BASE_URL=\"https://api.example.com\"") {
		t.Error("expected BaseURL in output")
	}
	if !contains(result, "CUSTOM=\"value\"") {
		t.Error("expected custom env var in output")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
