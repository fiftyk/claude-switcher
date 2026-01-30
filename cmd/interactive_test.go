package cmd

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/fiftyk/claude-switcher/internal/profile"
)

// mockMenuHandler 用于测试
type mockMenuHandler struct {
	showMenuCalled  bool
	setActiveCalled bool
	shouldFail      bool
	returnAction    MenuAction
	returnName      string
	returnErr       error
}

func (m *mockMenuHandler) ShowMenu(profilesDir string) (MenuAction, string, error) {
	m.showMenuCalled = true
	return m.returnAction, m.returnName, m.returnErr
}

func (m *mockMenuHandler) SetActiveProfile(name string) error {
	m.setActiveCalled = true
	if m.shouldFail {
		return errors.New("mock error")
	}
	return nil
}

// TestHandleMenuAction_Quit 测试退出操作
func TestHandleMenuAction_Quit(t *testing.T) {
	// Mock RunClaude 和 SyncToSettings
	runClaudeFunc = func() error { return nil }
	syncToSettingsFunc = func(profileName string, p *profile.Profile) error { return nil }
	defer func() {
		runClaudeFunc = defaultRunClaude
		syncToSettingsFunc = defaultSyncToSettings
	}()

	profilesDir := t.TempDir()
	handler := &mockMenuHandler{
		returnAction: ActionQuit,
	}

	err := HandleMenuAction(profilesDir, ActionQuit, "", handler)
	if err != ErrQuit {
		t.Errorf("expected ErrQuit, got %v", err)
	}
}

// TestHandleMenuAction_Run 测试运行操作
func TestHandleMenuAction_Run(t *testing.T) {
	// Mock RunClaude 和 SyncToSettings
	syncCalled := false
	runClaudeFunc = func() error { return nil }
	syncToSettingsFunc = func(profileName string, p *profile.Profile) error { syncCalled = true; return nil }
	defer func() {
		runClaudeFunc = defaultRunClaude
		syncToSettingsFunc = defaultSyncToSettings
	}()

	profilesDir := t.TempDir()

	// 创建测试配置
	p := &profile.Profile{
		Name:       "test",
		AuthToken:  "sk-test",
		BaseURL:    "https://api.example.com",
		HTTPProxy:  "http://127.0.0.1:7890",
		HTTPSProxy: "http://127.0.0.1:7890",
	}
	content := formatProfile(p)
	if err := os.WriteFile(filepath.Join(profilesDir, "test.conf"), []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	handler := &mockMenuHandler{
		returnAction: ActionRun,
		returnName:   "test",
	}

	err := HandleMenuAction(profilesDir, ActionRun, "test", handler)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !handler.setActiveCalled {
		t.Error("expected SetActiveProfile to be called")
	}
	if !syncCalled {
		t.Error("expected SyncToSettings to be called")
	}
}

// TestHandleMenuAction_Edit 测试编辑操作
func TestHandleMenuAction_Edit(t *testing.T) {
	profilesDir := t.TempDir()

	// 创建测试配置
	p := &profile.Profile{
		Name:       "test",
		AuthToken:  "sk-test",
	}
	content := formatProfile(p)
	if err := os.WriteFile(filepath.Join(profilesDir, "test.conf"), []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	// 验证 EditProfileInteractive 函数存在
	// 注意：交互式编辑需要 stdin，此测试仅验证函数签名正确
	err := EditProfileInteractive(profilesDir, "test")
	if err != nil {
		// stdin 为空时的预期行为
		if err.Error() != "EOF" && err.Error() != "unexpected EOF" {
			t.Errorf("unexpected error: %v", err)
		}
	}
}

// TestHandleMenuAction_Delete 测试删除操作
func TestHandleMenuAction_Delete(t *testing.T) {
	profilesDir := t.TempDir()

	// 创建测试配置
	p := &profile.Profile{
		Name:       "test",
		AuthToken:  "sk-test",
	}
	content := formatProfile(p)
	if err := os.WriteFile(filepath.Join(profilesDir, "test.conf"), []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	// 验证 DeleteProfileInteractive 函数存在且能被调用
	// 注意：交互式删除需要 stdin，此测试仅验证函数签名正确
	err := DeleteProfileInteractive(profilesDir, "test")
	if err != nil {
		// 用户取消删除是预期行为（因为 stdin 为空）
		if err.Error() != "canceled" {
			t.Errorf("unexpected error: %v", err)
		}
	}

	// 验证配置文件仍然存在（因为用户取消了删除）
	if _, err := os.Stat(filepath.Join(profilesDir, "test.conf")); os.IsNotExist(err) {
		t.Error("profile should not be deleted when user cancels")
	}
}

// TestHandleMenuAction_ShowDetails 测试显示详情操作
func TestHandleMenuAction_ShowDetails(t *testing.T) {
	profilesDir := t.TempDir()

	// 创建测试配置
	p := &profile.Profile{
		Name:       "test",
		AuthToken:  "sk-test",
		BaseURL:    "https://api.example.com",
		HTTPProxy:  "http://127.0.0.1:7890",
		Model:      "claude-3-5-sonnet",
		EnvVars:    map[string]string{"TEST": "value"},
	}
	content := formatProfile(p)
	if err := os.WriteFile(filepath.Join(profilesDir, "test.conf"), []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	// 验证 ShowProfileDetails 函数存在且正确解析配置
	err := ShowProfileDetails(profilesDir, "test")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestHandleMenuAction_Export 测试导出操作
func TestHandleMenuAction_Export(t *testing.T) {
	profilesDir := t.TempDir()

	// 创建测试配置
	p := &profile.Profile{
		Name:       "test",
		AuthToken:  "sk-test",
		BaseURL:    "https://api.example.com",
	}
	content := formatProfile(p)
	if err := os.WriteFile(filepath.Join(profilesDir, "test.conf"), []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	// 验证 ExportProfileWithFormat 函数可用
	outputPath := filepath.Join(profilesDir, "exported.json")
	err := ExportProfileWithFormat(profilesDir, "test", FormatJSON, outputPath)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// 验证导出文件存在
	if _, err := os.Stat(outputPath); err != nil {
		t.Error("expected exported file to exist")
	}
}

// TestHandleMenuAction_LoadProfileError 测试加载配置错误
func TestHandleMenuAction_LoadProfileError(t *testing.T) {
	// Mock SyncToSettings
	syncToSettingsFunc = func(profileName string, p *profile.Profile) error { return nil }
	defer func() { syncToSettingsFunc = defaultSyncToSettings }()

	profilesDir := t.TempDir()
	handler := &mockMenuHandler{
		returnAction: ActionRun,
		returnName:   "nonexistent",
	}

	err := HandleMenuAction(profilesDir, ActionRun, "nonexistent", handler)
	if err == nil {
		t.Error("expected error for nonexistent profile")
	}
}

// TestHandleMenuAction_SetActiveProfileError 测试设置活动配置错误
func TestHandleMenuAction_SetActiveProfileError(t *testing.T) {
	// Mock SyncToSettings
	syncToSettingsFunc = func(profileName string, p *profile.Profile) error { return nil }
	defer func() { syncToSettingsFunc = defaultSyncToSettings }()

	profilesDir := t.TempDir()

	// 创建测试配置
	p := &profile.Profile{
		Name:      "test",
		AuthToken: "sk-test",
	}
	content := formatProfile(p)
	if err := os.WriteFile(filepath.Join(profilesDir, "test.conf"), []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	handler := &mockMenuHandler{
		returnAction: ActionRun,
		returnName:   "test",
		shouldFail:   true,
	}

	err := HandleMenuAction(profilesDir, ActionRun, "test", handler)
	if err == nil {
		t.Error("expected error when SetActiveProfile fails")
	}
}
