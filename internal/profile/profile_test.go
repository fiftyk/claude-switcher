package profile

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadProfile(t *testing.T) {
	// 创建临时配置文件
	tmpDir := t.TempDir()
	profileName := "test-profile"
	profileFile := filepath.Join(tmpDir, profileName+".conf")

	// 写入测试配置
	content := `# Claude Switcher 配置文件
NAME="测试配置"
ANTHROPIC_AUTH_TOKEN="sk-test-token"
ANTHROPIC_BASE_URL="https://api.example.com"
http_proxy="http://127.0.0.1:7890"
ANTHROPIC_MODEL="claude-3-5-sonnet"
`
	if err := os.WriteFile(profileFile, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	// 加载配置
	p, err := LoadProfile(tmpDir, profileName)
	if err != nil {
		t.Fatalf("LoadProfile() error = %v", err)
	}

	// 验证
	if p.Name != "测试配置" {
		t.Errorf("Name = %v, want %v", p.Name, "测试配置")
	}
	if p.AuthToken != "sk-test-token" {
		t.Errorf("AuthToken = %v, want %v", p.AuthToken, "sk-test-token")
	}
	if p.BaseURL != "https://api.example.com" {
		t.Errorf("BaseURL = %v, want %v", p.BaseURL, "https://api.example.com")
	}
	if p.HTTPProxy != "http://127.0.0.1:7890" {
		t.Errorf("HTTPProxy = %v, want %v", p.HTTPProxy, "http://127.0.0.1:7890")
	}
	if p.HTTPSProxy != "http://127.0.0.1:7890" {
		t.Errorf("HTTPSProxy = %v, want %v", p.HTTPSProxy, "http://127.0.0.1:7890")
	}
	if p.Model != "claude-3-5-sonnet" {
		t.Errorf("Model = %v, want %v", p.Model, "claude-3-5-sonnet")
	}
}

func TestLoadProfileNotFound(t *testing.T) {
	_, err := LoadProfile("/nonexistent", "test")
	if err == nil {
		t.Error("LoadProfile() expected error for non-existent profile")
	}
}

func TestLoadProfileWithCustomEnvVars(t *testing.T) {
	tmpDir := t.TempDir()
	profileName := "custom-env"
	profileFile := filepath.Join(tmpDir, profileName+".conf")

	content := `NAME="Custom Env"
ANTHROPIC_AUTH_TOKEN="sk-token"
CUSTOM_VAR="custom-value"
ANOTHER_VAR="another"
`
	if err := os.WriteFile(profileFile, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	p, err := LoadProfile(tmpDir, profileName)
	if err != nil {
		t.Fatal(err)
	}

	// 验证自定义环境变量
	if p.EnvVars["CUSTOM_VAR"] != "custom-value" {
		t.Errorf("EnvVars[CUSTOM_VAR] = %v, want %v", p.EnvVars["CUSTOM_VAR"], "custom-value")
	}
	if p.EnvVars["ANOTHER_VAR"] != "another" {
		t.Errorf("EnvVars[ANOTHER_VAR] = %v, want %v", p.EnvVars["ANOTHER_VAR"], "another")
	}
}

func TestLoadProfileSkipsCommentsAndEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	profileName := "skip-test"
	profileFile := filepath.Join(tmpDir, profileName+".conf")

	content := `# 这是注释
NAME="Skip Test"

# 空行应该被跳过
ANTHROPIC_AUTH_TOKEN="sk-token"
`
	if err := os.WriteFile(profileFile, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	p, err := LoadProfile(tmpDir, profileName)
	if err != nil {
		t.Fatal(err)
	}

	if p.AuthToken != "sk-token" {
		t.Errorf("AuthToken = %v, want %v", p.AuthToken, "sk-token")
	}
	// NAME 变量不应出现在 EnvVars 中（由 Name 字段处理）
	if _, ok := p.EnvVars["NAME"]; ok {
		t.Error("NAME should not be in EnvVars")
	}
}

func TestListProfiles(t *testing.T) {
	tmpDir := t.TempDir()

	// 创建多个配置文件
	profiles := []string{"profile1", "profile2", "profile3"}
	for _, name := range profiles {
		content := `NAME="` + name + `"
ANTHROPIC_AUTH_TOKEN="sk-token"
`
		if err := os.WriteFile(filepath.Join(tmpDir, name+".conf"), []byte(content), 0600); err != nil {
			t.Fatal(err)
		}
	}

	names, err := ListProfiles(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	if len(names) != 3 {
		t.Errorf("ListProfiles() returned %d profiles, want 3", len(names))
	}
}

func TestListProfilesEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	names, err := ListProfiles(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(names) != 0 {
		t.Errorf("ListProfiles() returned %d profiles, want 0", len(names))
	}
}

func TestDeleteProfile(t *testing.T) {
	tmpDir := t.TempDir()
	profileName := "to-delete"
	profileFile := filepath.Join(tmpDir, profileName+".conf")

	content := `NAME="To Delete"
ANTHROPIC_AUTH_TOKEN="sk-token"
`
	if err := os.WriteFile(profileFile, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	if err := DeleteProfile(tmpDir, profileName); err != nil {
		t.Fatalf("DeleteProfile() error = %v", err)
	}

	if _, err := os.Stat(profileFile); !os.IsNotExist(err) {
		t.Error("Profile file was not deleted")
	}
}

func TestDeleteProfileNotFound(t *testing.T) {
	err := DeleteProfile("/nonexistent", "test")
	if err == nil {
		t.Error("DeleteProfile() expected error for non-existent profile")
	}
}

func TestCopyProfile(t *testing.T) {
	tmpDir := t.TempDir()
	srcName := "source"
	dstName := "destination"

	// 创建源配置
	srcFile := filepath.Join(tmpDir, srcName+".conf")
	content := `NAME="Source Profile"
ANTHROPIC_AUTH_TOKEN="sk-source-token"
ANTHROPIC_BASE_URL="https://api.source.com"
`
	if err := os.WriteFile(srcFile, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	if err := CopyProfile(tmpDir, srcName, dstName); err != nil {
		t.Fatalf("CopyProfile() error = %v", err)
	}

	// 验证目标文件存在
	dstFile := filepath.Join(tmpDir, dstName+".conf")
	if _, err := os.Stat(dstFile); os.IsNotExist(err) {
		t.Error("Destination profile file was not created")
	}

	// 验证内容
	p, err := LoadProfile(tmpDir, dstName)
	if err != nil {
		t.Fatal(err)
	}
	if p.AuthToken != "sk-source-token" {
		t.Errorf("AuthToken = %v, want %v", p.AuthToken, "sk-source-token")
	}
	if p.Name != dstName {
		t.Errorf("Name = %v, want %v", p.Name, dstName)
	}
}

func TestRenameProfile(t *testing.T) {
	tmpDir := t.TempDir()
	oldName := "old-name"
	newName := "new-name"

	// 创建旧配置
	oldFile := filepath.Join(tmpDir, oldName+".conf")
	content := `NAME="Old Name"
ANTHROPIC_AUTH_TOKEN="sk-token"
`
	if err := os.WriteFile(oldFile, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	if err := RenameProfile(tmpDir, oldName, newName); err != nil {
		t.Fatalf("RenameProfile() error = %v", err)
	}

	// 验证旧文件不存在
	if _, err := os.Stat(oldFile); !os.IsNotExist(err) {
		t.Error("Old profile file still exists")
	}

	// 验证新文件存在
	newFile := filepath.Join(tmpDir, newName+".conf")
	if _, err := os.Stat(newFile); os.IsNotExist(err) {
		t.Error("New profile file was not created")
	}
}
