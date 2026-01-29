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
	showMenuCalled   bool
	setActiveCalled  bool
	runClaudeCalled  bool
	shouldFail       bool
	returnAction     MenuAction
	returnName       string
	returnErr        error
}

func (m *mockMenuHandler) ShowMenu(profilesDir string) (MenuAction, string, error) {
	m.showMenuCalled = true
	return m.returnAction, m.returnName, m.returnErr
}

func (m *mockMenuHandler) RunClaude(p *profile.Profile, args ...string) error {
	m.runClaudeCalled = true
	if m.shouldFail {
		return errors.New("mock error")
	}
	return nil
}

func (m *mockMenuHandler) SetActiveProfile(name string) error {
	m.setActiveCalled = true
	if m.shouldFail {
		return errors.New("mock error")
	}
	return nil
}

// mockSettingsSyncer 用于测试
type mockSettingsSyncer struct {
	syncCalled bool
	shouldFail bool
}

func (s *mockSettingsSyncer) SyncToSettings(profileName string, p *profile.Profile) error {
	s.syncCalled = true
	if s.shouldFail {
		return errors.New("mock sync error")
	}
	return nil
}

// noopSettingsSyncer 不进行实际同步的 syncer（用于测试）
type noopSettingsSyncer struct{}

func (s noopSettingsSyncer) SyncToSettings(profileName string, p *profile.Profile) error {
	return nil
}

func TestHandleMenuAction_Quit(t *testing.T) {
	profilesDir := t.TempDir()
	handler := &mockMenuHandler{
		returnAction: ActionQuit,
	}
	syncer := noopSettingsSyncer{}

	err := HandleMenuAction(profilesDir, ActionQuit, "", handler, syncer)
	if err != ErrQuit {
		t.Errorf("expected ErrQuit, got %v", err)
	}
}

func TestHandleMenuAction_Run(t *testing.T) {
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
	syncer := noopSettingsSyncer{}

	err := HandleMenuAction(profilesDir, ActionRun, "test", handler, syncer)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !handler.setActiveCalled {
		t.Error("expected SetActiveProfile to be called")
	}
	if !handler.runClaudeCalled {
		t.Error("expected RunClaude to be called")
	}
}

func TestHandleMenuAction_RunWithSync(t *testing.T) {
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

	handler := &mockMenuHandler{
		returnAction: ActionRunWithSync,
		returnName:   "test",
	}
	syncer := &mockSettingsSyncer{}

	err := HandleMenuAction(profilesDir, ActionRunWithSync, "test", handler, syncer)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !handler.setActiveCalled {
		t.Error("expected SetActiveProfile to be called")
	}
	if !handler.runClaudeCalled {
		t.Error("expected RunClaude to be called")
	}
	if !syncer.syncCalled {
		t.Error("expected SyncToSettings to be called")
	}
}

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

func TestHandleMenuAction_LoadProfileError(t *testing.T) {
	profilesDir := t.TempDir()
	handler := &mockMenuHandler{
		returnAction: ActionRun,
		returnName:   "nonexistent",
	}
	syncer := noopSettingsSyncer{}

	err := HandleMenuAction(profilesDir, ActionRun, "nonexistent", handler, syncer)
	if err == nil {
		t.Error("expected error for nonexistent profile")
	}
}

func TestHandleMenuAction_SetActiveProfileError(t *testing.T) {
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
	syncer := noopSettingsSyncer{}

	err := HandleMenuAction(profilesDir, ActionRun, "test", handler, syncer)
	if err == nil {
		t.Error("expected error when SetActiveProfile fails")
	}
}

func TestHandleMenuAction_SyncError(t *testing.T) {
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
		returnAction: ActionRunWithSync,
		returnName:   "test",
	}
	syncer := &mockSettingsSyncer{shouldFail: true}

	err := HandleMenuAction(profilesDir, ActionRunWithSync, "test", handler, syncer)
	if err == nil {
		t.Error("expected error when SyncToSettings fails")
	}
}
