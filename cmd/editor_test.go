package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fiftyk/claude-switcher/internal/profile"
)

// mockReader 用于测试
type mockReader struct {
	inputs []string
	pos    int
}

func (r *mockReader) ReadString(delim byte) (string, error) {
	if r.pos >= len(r.inputs) {
		return "", nil
	}
	input := r.inputs[r.pos]
	r.pos++
	return input + "\n", nil
}

func TestEditProfileInteractiveWithReader(t *testing.T) {
	profilesDir := t.TempDir()

	// 创建测试配置
	originalProfile := &profile.Profile{
		Name:       "test",
		AuthToken:  "sk-original-token",
		BaseURL:    "https://original.example.com",
		HTTPProxy:  "http://127.0.0.1:8080",
		HTTPSProxy: "http://127.0.0.1:8080",
		Model:      "claude-2",
	}
	content := formatProfile(originalProfile)
	if err := os.WriteFile(filepath.Join(profilesDir, "test.conf"), []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	// 创建 mock reader，模拟用户输入（直接回车保持所有值）
	reader := &mockReader{
		inputs: []string{""}, // 所有输入都直接回车
	}

	err := EditProfileInteractiveWithReader(profilesDir, "test", reader)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// 验证配置文件未被修改
	p, err := profile.LoadProfile(profilesDir, "test")
	if err != nil {
		t.Fatal(err)
	}
	if p.AuthToken != "sk-original-token" {
		t.Errorf("expected AuthToken to remain unchanged, got %s", p.AuthToken)
	}
}

func TestEditProfileInteractiveWithReader_UpdateToken(t *testing.T) {
	profilesDir := t.TempDir()

	// 创建测试配置
	originalProfile := &profile.Profile{
		Name:      "test",
		AuthToken: "sk-original",
	}
	content := formatProfile(originalProfile)
	if err := os.WriteFile(filepath.Join(profilesDir, "test.conf"), []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	// 创建 mock reader，更新 token
	reader := &mockReader{
		inputs: []string{"", "sk-new-token", "", "", "", ""}, // 只更新 token
	}

	err := EditProfileInteractiveWithReader(profilesDir, "test", reader)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// 验证 token 已更新
	p, err := profile.LoadProfile(profilesDir, "test")
	if err != nil {
		t.Fatal(err)
	}
	if p.AuthToken != "sk-new-token" {
		t.Errorf("expected AuthToken to be sk-new-token, got %s", p.AuthToken)
	}
}

func TestEditProfileInteractive_ProfileNotFound(t *testing.T) {
	profilesDir := t.TempDir()
	reader := &mockReader{}

	err := EditProfileInteractiveWithReader(profilesDir, "nonexistent", reader)
	if err == nil {
		t.Error("expected error for nonexistent profile")
	}
}

func TestMaskValue(t *testing.T) {
	tests := []struct {
		name   string
		value  string
		expect string
	}{
		{"normal value", "hello world", "he*******ld"},
		{"short value", "hi", "****"},
		{"empty value", "", "****"},
		{"4 chars", "test", "****"}, // 长度<=4时完全遮盖
		{"5 chars", "hello", "he*lo"},
		{"6 chars", "world!", "wo**d!"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := maskValue(tt.value)
			if result != tt.expect {
				t.Errorf("maskValue(%s) = %s, want %s", tt.value, result, tt.expect)
			}
		})
	}
}

func TestFormatProfile_AllFields(t *testing.T) {
	p := &profile.Profile{
		Name:       "Test Profile",
		AuthToken:  "sk-test-token",
		BaseURL:    "https://api.example.com",
		HTTPProxy:  "http://127.0.0.1:7890",
		HTTPSProxy: "http://127.0.0.1:7890",
		Model:      "claude-3-5-sonnet",
		EnvVars:    map[string]string{"CUSTOM_VAR": "custom_value"},
	}

	result := formatProfile(p)

	// 验证所有字段都被格式化
	if !strings.Contains(result, "NAME=\"Test Profile\"") {
		t.Error("expected NAME in output")
	}
	if !strings.Contains(result, "ANTHROPIC_AUTH_TOKEN=\"sk-test-token\"") {
		t.Error("expected AuthToken in output")
	}
	if !strings.Contains(result, "ANTHROPIC_BASE_URL=\"https://api.example.com\"") {
		t.Error("expected BaseURL in output")
	}
	if !strings.Contains(result, "http_proxy=\"http://127.0.0.1:7890\"") {
		t.Error("expected http_proxy in output")
	}
	if !strings.Contains(result, "ANTHROPIC_MODEL=\"claude-3-5-sonnet\"") {
		t.Error("expected Model in output")
	}
	if !strings.Contains(result, "CUSTOM_VAR=\"custom_value\"") {
		t.Error("expected custom env var in output")
	}
}

func TestFormatProfile_EmptyFields(t *testing.T) {
	p := &profile.Profile{
		Name: "Minimal",
	}

	result := formatProfile(p)

	// 验证包含 NAME
	if !strings.Contains(result, "NAME=\"Minimal\"") {
		t.Error("expected NAME in output")
	}
	// 验证只有2行（注释 + NAME）
	lines := strings.Split(strings.TrimSpace(result), "\n")
	if len(lines) != 2 {
		t.Errorf("expected 2 lines (comment + NAME), got %d", len(lines))
	}
}
