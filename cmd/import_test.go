package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestImportProfileFromJSON(t *testing.T) {
	// 模拟 JSON 导入数据
	jsonData := `{
  "name": "Imported Profile",
  "auth_token": "sk-imported-token",
  "base_url": "https://api.imported.com",
  "http_proxy": "http://127.0.0.1:7890",
  "model": "claude-3-5-sonnet",
  "custom_vars": {
    "CUSTOM_VAR": "custom-value"
  }
}`

	// 导入配置
	profile, err := ImportProfileFromJSON(jsonData, "imported")
	if err != nil {
		t.Fatalf("ImportProfileFromJSON failed: %v", err)
	}

	// 验证
	if profile.Name != "Imported Profile" {
		t.Errorf("expected name to be 'Imported Profile', got %s", profile.Name)
	}
	if profile.AuthToken != "sk-imported-token" {
		t.Errorf("expected AuthToken to be 'sk-imported-token', got %s", profile.AuthToken)
	}
	if profile.BaseURL != "https://api.imported.com" {
		t.Errorf("expected BaseURL to be 'https://api.imported.com', got %s", profile.BaseURL)
	}
	if profile.HTTPProxy != "http://127.0.0.1:7890" {
		t.Errorf("expected HTTPProxy to be 'http://127.0.0.1:7890', got %s", profile.HTTPProxy)
	}
	if profile.Model != "claude-3-5-sonnet" {
		t.Errorf("expected Model to be 'claude-3-5-sonnet', got %s", profile.Model)
	}
	if profile.EnvVars["CUSTOM_VAR"] != "custom-value" {
		t.Errorf("expected CUSTOM_VAR to be 'custom-value', got %s", profile.EnvVars["CUSTOM_VAR"])
	}
}

func TestImportProfileFromYAML(t *testing.T) {
	// 模拟 YAML 导入数据
	yamlData := `name: YAML Profile
auth_token: "sk-yaml-token"
base_url: "https://api.yaml.com"
http_proxy: "http://127.0.0.1:1080"
model: "claude-3-haiku"
custom_vars:
  YAML_VAR: "yaml-value"
`

	// 导入配置
	profile, err := ImportProfileFromYAML(yamlData, "yaml-profile")
	if err != nil {
		t.Fatalf("ImportProfileFromYAML failed: %v", err)
	}

	// 验证
	if profile.Name != "YAML Profile" {
		t.Errorf("expected name to be 'YAML Profile', got %s", profile.Name)
	}
	if profile.AuthToken != "sk-yaml-token" {
		t.Errorf("expected AuthToken to be 'sk-yaml-token', got %s", profile.AuthToken)
	}
	if profile.BaseURL != "https://api.yaml.com" {
		t.Errorf("expected BaseURL to be 'https://api.yaml.com', got %s", profile.BaseURL)
	}
	if profile.EnvVars["YAML_VAR"] != "yaml-value" {
		t.Errorf("expected YAML_VAR to be 'yaml-value', got %s", profile.EnvVars["YAML_VAR"])
	}
}

func TestImportProfileInvalidJSON(t *testing.T) {
	// 测试无效 JSON
	_, err := ImportProfileFromJSON("invalid json", "test")
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestImportProfileInvalidYAML(t *testing.T) {
	// 测试无效 YAML
	_, err := ImportProfileFromYAML("invalid: yaml: content: [", "test")
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}

func TestSaveImportedProfile(t *testing.T) {
	// 创建临时目录
	tmpDir := t.TempDir()

	// 导入并保存
	jsonData := `{
  "name": "Save Test Profile",
  "auth_token": "sk-save-token"
}`

	profile, err := ImportProfileFromJSON(jsonData, "save-test")
	if err != nil {
		t.Fatal(err)
	}

	// 保存到文件，使用 "save-test" 作为文件名
	err = SaveProfileToFile(tmpDir, profile, "save-test")
	if err != nil {
		t.Fatalf("SaveProfileToFile failed: %v", err)
	}

	// 验证文件存在
	filePath := filepath.Join(tmpDir, "save-test.conf")
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Error("profile file was not created")
	}

	// 验证文件内容
	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatal(err)
	}

	if !contains(string(data), "NAME=\"Save Test Profile\"") {
		t.Error("saved profile should contain NAME")
	}
	if !contains(string(data), "ANTHROPIC_AUTH_TOKEN=\"sk-save-token\"") {
		t.Error("saved profile should contain AuthToken")
	}
}
