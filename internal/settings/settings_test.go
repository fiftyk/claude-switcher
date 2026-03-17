package settings

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadSettings(t *testing.T) {
	// 创建临时 settings.json
	tmpDir := t.TempDir()
	settingsFile := filepath.Join(tmpDir, "settings.json")

	content := `{
  "enabledPlugins": {
    "test-plugin": true
  },
  "env": {
    "ANTHROPIC_API_KEY": "sk-test"
  },
  "_claudeSwitcherProfile": "test-profile"
}`
	if err := os.WriteFile(settingsFile, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	s, err := LoadSettings(settingsFile)
	if err != nil {
		t.Fatalf("LoadSettings() error = %v", err)
	}

	if s.ClaudeSwitcherProfile != "test-profile" {
		t.Errorf("ClaudeSwitcherProfile = %v, want %v", s.ClaudeSwitcherProfile, "test-profile")
	}
	if s.Env["ANTHROPIC_API_KEY"] != "sk-test" {
		t.Errorf("Env[ANTHROPIC_API_KEY] = %v, want %v", s.Env["ANTHROPIC_API_KEY"], "sk-test")
	}
	if !s.EnabledPlugins["test-plugin"] {
		t.Error("EnabledPlugins[test-plugin] should be true")
	}
}

func TestLoadSettingsNotFound(t *testing.T) {
	_, err := LoadSettings("/nonexistent/settings.json")
	if err == nil {
		t.Error("LoadSettings() expected error for non-existent file")
	}
}

func TestLoadSettingsEmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	settingsFile := filepath.Join(tmpDir, "settings.json")

	// 创建空文件
	if err := os.WriteFile(settingsFile, []byte(""), 0600); err != nil {
		t.Fatal(err)
	}

	_, err := LoadSettings(settingsFile)
	if err == nil {
		t.Error("LoadSettings() expected error for empty file")
	}
}

func TestSaveSettings(t *testing.T) {
	tmpDir := t.TempDir()
	settingsFile := filepath.Join(tmpDir, "settings.json")

	s := &Settings{
		Env: map[string]string{
			"ANTHROPIC_API_KEY": "sk-new-key",
		},
		ClaudeSwitcherProfile: "new-profile",
	}

	if err := SaveSettings(settingsFile, s); err != nil {
		t.Fatalf("SaveSettings() error = %v", err)
	}

	// 验证文件内容
	data, err := os.ReadFile(settingsFile)
	if err != nil {
		t.Fatal(err)
	}

	var saved Settings
	if err := json.Unmarshal(data, &saved); err != nil {
		t.Fatal(err)
	}

	if saved.Env["ANTHROPIC_API_KEY"] != "sk-new-key" {
		t.Errorf("saved.Env[ANTHROPIC_API_KEY] = %v, want %v", saved.Env["ANTHROPIC_API_KEY"], "sk-new-key")
	}
	if saved.ClaudeSwitcherProfile != "new-profile" {
		t.Errorf("saved.ClaudeSwitcherProfile = %v, want %v", saved.ClaudeSwitcherProfile, "new-profile")
	}
}

func TestSyncProfileToSettings(t *testing.T) {
	tmpDir := t.TempDir()
	settingsFile := filepath.Join(tmpDir, "settings.json")

	// 创建初始 settings.json
	initialContent := `{
  "enabledPlugins": {
    "test-plugin": true
  }
}`
	if err := os.WriteFile(settingsFile, []byte(initialContent), 0600); err != nil {
		t.Fatal(err)
	}

	envVars := map[string]string{
		"ANTHROPIC_API_KEY": "sk-sync-key",
		"ANTHROPIC_BASE_URL": "https://api.example.com",
	}

	if err := SyncProfileToSettings(settingsFile, "test-profile", envVars); err != nil {
		t.Fatalf("SyncProfileToSettings() error = %v", err)
	}

	// 验证
	s, err := LoadSettings(settingsFile)
	if err != nil {
		t.Fatal(err)
	}

	// 保留原有配置
	if !s.EnabledPlugins["test-plugin"] {
		t.Error("EnabledPlugins[test-plugin] should be preserved")
	}

	// 更新 env
	if s.Env["ANTHROPIC_API_KEY"] != "sk-sync-key" {
		t.Errorf("Env[ANTHROPIC_API_KEY] = %v, want %v", s.Env["ANTHROPIC_API_KEY"], "sk-sync-key")
	}
	if s.Env["ANTHROPIC_BASE_URL"] != "https://api.example.com" {
		t.Errorf("Env[ANTHROPIC_BASE_URL] = %v, want %v", s.Env["ANTHROPIC_BASE_URL"], "https://api.example.com")
	}

	// 设置 profile 标记
	if s.ClaudeSwitcherProfile != "test-profile" {
		t.Errorf("ClaudeSwitcherProfile = %v, want %v", s.ClaudeSwitcherProfile, "test-profile")
	}
}

func TestSyncProfileToSettingsNewFile(t *testing.T) {
	tmpDir := t.TempDir()
	settingsFile := filepath.Join(tmpDir, "settings.json")

	envVars := map[string]string{
		"ANTHROPIC_API_KEY": "sk-new",
	}

	if err := SyncProfileToSettings(settingsFile, "new-profile", envVars); err != nil {
		t.Fatalf("SyncProfileToSettings() error = %v", err)
	}

	// 验证文件存在
	if _, err := os.Stat(settingsFile); os.IsNotExist(err) {
		t.Error("settings.json was not created")
	}
}

// TestSyncProfileToSettingsClearsOldEnv 验证 sync 时清空旧的 env 变量
func TestSyncProfileToSettingsClearsOldEnv(t *testing.T) {
	tmpDir := t.TempDir()
	settingsFile := filepath.Join(tmpDir, "settings.json")

	// 创建包含旧环境变量的 settings.json
	initialContent := `{
  "enabledPlugins": {
    "test-plugin": true
  },
  "env": {
    "ANTHROPIC_AUTH_TOKEN": "old-token",
    "ANTHROPIC_BASE_URL": "https://old.api.com",
    "ANTHROPIC_DEFAULT_HAIKU_MODEL": "MiniMax-M2.5",
    "ANTHROPIC_DEFAULT_OPUS_MODEL": "MiniMax-M2.5",
    "ANTHROPIC_DEFAULT_SONNET_MODEL": "MiniMax-M2.5"
  },
  "_claudeSwitcherProfile": "old-profile"
}`
	if err := os.WriteFile(settingsFile, []byte(initialContent), 0600); err != nil {
		t.Fatal(err)
	}

	// 用新 profile 同步，只包含 token 和 base_url
	newEnvVars := map[string]string{
		"ANTHROPIC_AUTH_TOKEN": "new-token",
		"ANTHROPIC_BASE_URL":  "https://new.api.com",
	}

	if err := SyncProfileToSettings(settingsFile, "new-profile", newEnvVars); err != nil {
		t.Fatalf("SyncProfileToSettings() error = %v", err)
	}

	s, err := LoadSettings(settingsFile)
	if err != nil {
		t.Fatal(err)
	}

	// 新 profile 的变量应该存在
	if s.Env["ANTHROPIC_AUTH_TOKEN"] != "new-token" {
		t.Errorf("ANTHROPIC_AUTH_TOKEN = %v, want new-token", s.Env["ANTHROPIC_AUTH_TOKEN"])
	}
	if s.Env["ANTHROPIC_BASE_URL"] != "https://new.api.com" {
		t.Errorf("ANTHROPIC_BASE_URL = %v, want https://new.api.com", s.Env["ANTHROPIC_BASE_URL"])
	}

	// 旧 profile 残留的变量应该被清除
	if _, exists := s.Env["ANTHROPIC_DEFAULT_HAIKU_MODEL"]; exists {
		t.Error("ANTHROPIC_DEFAULT_HAIKU_MODEL should be cleared after sync")
	}
	if _, exists := s.Env["ANTHROPIC_DEFAULT_OPUS_MODEL"]; exists {
		t.Error("ANTHROPIC_DEFAULT_OPUS_MODEL should be cleared after sync")
	}
	if _, exists := s.Env["ANTHROPIC_DEFAULT_SONNET_MODEL"]; exists {
		t.Error("ANTHROPIC_DEFAULT_SONNET_MODEL should be cleared after sync")
	}

	// enabledPlugins 应该保留
	if !s.EnabledPlugins["test-plugin"] {
		t.Error("enabledPlugins should be preserved")
	}

	// profile 标记应该更新
	if s.ClaudeSwitcherProfile != "new-profile" {
		t.Errorf("ClaudeSwitcherProfile = %v, want new-profile", s.ClaudeSwitcherProfile)
	}
}

func TestSettingsFilePermissions(t *testing.T) {
	tmpDir := t.TempDir()
	settingsFile := filepath.Join(tmpDir, "settings.json")

	s := &Settings{
		Env: map[string]string{
			"TEST": "value",
		},
	}

	if err := SaveSettings(settingsFile, s); err != nil {
		t.Fatalf("SaveSettings() error = %v", err)
	}

	info, err := os.Stat(settingsFile)
	if err != nil {
		t.Fatal(err)
	}

	// 检查权限 (0600 = -rw-------)
	perm := info.Mode().Perm()
	if perm != 0600 {
		t.Errorf("File permissions = %v, want %v", perm, 0600)
	}
}

func TestClearProfileEnvVars(t *testing.T) {
	tmpDir := t.TempDir()
	settingsFile := filepath.Join(tmpDir, "settings.json")

	// 创建初始 settings.json，包含 profile 环境变量
	initialContent := `{
  "enabledPlugins": {
    "test-plugin": true
  },
  "env": {
    "ANTHROPIC_AUTH_TOKEN": "sk-old-token",
    "ANTHROPIC_BASE_URL": "https://api.old.com",
    "http_proxy": "http://old-proxy:7890",
    "https_proxy": "http://old-proxy:7890",
    "ANTHROPIC_MODEL": "claude-3-5-sonnet",
    "CUSTOM_VAR": "custom-value"
  },
  "_claudeSwitcherProfile": "old-profile"
}`
	if err := os.WriteFile(settingsFile, []byte(initialContent), 0600); err != nil {
		t.Fatal(err)
	}

	// 清除 profile 环境变量
	if err := ClearProfileEnvVars(settingsFile, "old-profile"); err != nil {
		t.Fatalf("ClearProfileEnvVars() error = %v", err)
	}

	// 验证
	s, err := LoadSettings(settingsFile)
	if err != nil {
		t.Fatal(err)
	}

	// 应该清除的字段
	if s.Env["ANTHROPIC_AUTH_TOKEN"] != "" {
		t.Errorf("ANTHROPIC_AUTH_TOKEN should be cleared, got %v", s.Env["ANTHROPIC_AUTH_TOKEN"])
	}
	if s.Env["ANTHROPIC_BASE_URL"] != "" {
		t.Errorf("ANTHROPIC_BASE_URL should be cleared, got %v", s.Env["ANTHROPIC_BASE_URL"])
	}
	if s.Env["http_proxy"] != "" {
		t.Errorf("http_proxy should be cleared, got %v", s.Env["http_proxy"])
	}
	if s.Env["https_proxy"] != "" {
		t.Errorf("https_proxy should be cleared, got %v", s.Env["https_proxy"])
	}
	if s.Env["ANTHROPIC_MODEL"] != "" {
		t.Errorf("ANTHROPIC_MODEL should be cleared, got %v", s.Env["ANTHROPIC_MODEL"])
	}

	// 应该保留的字段
	if s.EnabledPlugins["test-plugin"] != true {
		t.Error("enabledPlugins should be preserved")
	}
	if s.Env["CUSTOM_VAR"] != "custom-value" {
		t.Errorf("CUSTOM_VAR should be preserved, got %v", s.Env["CUSTOM_VAR"])
	}
	if s.ClaudeSwitcherProfile != "" {
		t.Errorf("ClaudeSwitcherProfile should be cleared, got %v", s.ClaudeSwitcherProfile)
	}
}

func TestClearProfileEnvVarsFileNotExist(t *testing.T) {
	tmpDir := t.TempDir()
	settingsFile := filepath.Join(tmpDir, "nonexistent.json")

	// 文件不存在时应该不报错
	if err := ClearProfileEnvVars(settingsFile, "test-profile"); err != nil {
		t.Errorf("ClearProfileEnvVars() error = %v, want nil", err)
	}
}

// TestSyncProfileToSettingsPreservesUnknownFields 验证 sync 时保留未知字段
func TestSyncProfileToSettingsPreservesUnknownFields(t *testing.T) {
	tmpDir := t.TempDir()
	settingsFile := filepath.Join(tmpDir, "settings.json")

	initialContent := `{
  "env": {
    "OLD_VAR": "old-value"
  },
  "enabledPlugins": {
    "test-plugin": true
  },
  "_claudeSwitcherProfile": "old-profile",
  "skipDangerousModePermissionPrompt": true,
  "someOtherSetting": "value"
}`
	if err := os.WriteFile(settingsFile, []byte(initialContent), 0600); err != nil {
		t.Fatal(err)
	}

	envVars := map[string]string{
		"NEW_VAR": "new-value",
	}

	if err := SyncProfileToSettings(settingsFile, "new-profile", envVars); err != nil {
		t.Fatalf("SyncProfileToSettings() error = %v", err)
	}

	// 读取原始 JSON 验证未知字段被保留
	data, err := os.ReadFile(settingsFile)
	if err != nil {
		t.Fatal(err)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}

	if _, exists := raw["skipDangerousModePermissionPrompt"]; !exists {
		t.Error("skipDangerousModePermissionPrompt should be preserved")
	}
	if _, exists := raw["someOtherSetting"]; !exists {
		t.Error("someOtherSetting should be preserved")
	}

	// 已知字段也应该正确
	s, err := LoadSettings(settingsFile)
	if err != nil {
		t.Fatal(err)
	}
	if s.Env["NEW_VAR"] != "new-value" {
		t.Errorf("Env[NEW_VAR] = %v, want new-value", s.Env["NEW_VAR"])
	}
	if !s.EnabledPlugins["test-plugin"] {
		t.Error("enabledPlugins should be preserved")
	}
	if s.ClaudeSwitcherProfile != "new-profile" {
		t.Errorf("ClaudeSwitcherProfile = %v, want new-profile", s.ClaudeSwitcherProfile)
	}
}
