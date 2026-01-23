package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/fiftyk/claude-switcher/internal/profile"
)

func TestExportProfileToJSON(t *testing.T) {
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

	// 导出为 JSON
	jsonData, err := ExportProfileToJSON(p)
	if err != nil {
		t.Fatalf("ExportProfileToJSON failed: %v", err)
	}

	// 验证 JSON 格式正确
	var exported map[string]interface{}
	if err := json.Unmarshal(jsonData, &exported); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}

	// 验证字段
	if exported["name"] != "Test Profile" {
		t.Errorf("expected name to be 'Test Profile', got %v", exported["name"])
	}
	if exported["auth_token"] != "sk-test-token" {
		t.Errorf("expected auth_token to be 'sk-test-token', got %v", exported["auth_token"])
	}
	if exported["base_url"] != "https://api.example.com" {
		t.Errorf("expected base_url to be 'https://api.example.com', got %v", exported["base_url"])
	}
}

func TestExportProfileToYAML(t *testing.T) {
	// 创建临时目录和测试配置
	tmpDir := t.TempDir()
	profileName := "test-profile"
	profileFile := filepath.Join(tmpDir, profileName+".conf")

	content := `NAME="Test Profile"
ANTHROPIC_AUTH_TOKEN="sk-test-token"
ANTHROPIC_BASE_URL="https://api.example.com"
`
	if err := os.WriteFile(profileFile, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	// 加载配置
	p, err := profile.LoadProfile(tmpDir, profileName)
	if err != nil {
		t.Fatal(err)
	}

	// 导出为 YAML
	yamlData, err := ExportProfileToYAML(p)
	if err != nil {
		t.Fatalf("ExportProfileToYAML failed: %v", err)
	}

	// 验证 YAML 包含必要内容
	if len(yamlData) == 0 {
		t.Error("YAML output is empty")
	}

	// 验证包含配置名称
	if !contains(string(yamlData), "name: Test Profile") {
		t.Error("YAML output should contain 'name: Test Profile'")
	}
}

func TestExportProfileToShell(t *testing.T) {
	// 创建临时目录和测试配置
	tmpDir := t.TempDir()
	profileName := "test-profile"
	profileFile := filepath.Join(tmpDir, profileName+".conf")

	content := `NAME="Test Profile"
ANTHROPIC_AUTH_TOKEN="sk-test-token"
ANTHROPIC_BASE_URL="https://api.example.com"
http_proxy="http://127.0.0.1:7890"
`
	if err := os.WriteFile(profileFile, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	// 加载配置
	p, err := profile.LoadProfile(tmpDir, profileName)
	if err != nil {
		t.Fatal(err)
	}

	// 导出为 Shell
	shellData, err := ExportProfileToShell(p)
	if err != nil {
		t.Fatalf("ExportProfileToShell failed: %v", err)
	}

	// 验证 Shell 包含必要内容
	if !contains(string(shellData), "ANTHROPIC_AUTH_TOKEN=") {
		t.Error("Shell output should contain 'ANTHROPIC_AUTH_TOKEN='")
	}
	if !contains(string(shellData), "http_proxy=") {
		t.Error("Shell output should contain 'http_proxy='")
	}
}

func TestExportAllProfiles(t *testing.T) {
	// 创建临时目录和多个测试配置
	tmpDir := t.TempDir()

	profiles := []string{"profile1", "profile2", "profile3"}
	for _, name := range profiles {
		content := "NAME=\"" + name + "\"\nANTHROPIC_AUTH_TOKEN=\"sk-test\"\n"
		if err := os.WriteFile(filepath.Join(tmpDir, name+".conf"), []byte(content), 0600); err != nil {
			t.Fatal(err)
		}
	}

	// 导出所有配置
	jsonData, err := ExportAllProfiles(tmpDir)
	if err != nil {
		t.Fatalf("ExportAllProfiles failed: %v", err)
	}

	// 验证 JSON 格式正确
	var exported map[string]interface{}
	if err := json.Unmarshal(jsonData, &exported); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}

	// 验证包含所有配置
	profilesList, ok := exported["profiles"].([]interface{})
	if !ok {
		t.Fatal("exported JSON should contain 'profiles' array")
	}
	if len(profilesList) != 3 {
		t.Errorf("expected 3 profiles, got %d", len(profilesList))
	}
}
