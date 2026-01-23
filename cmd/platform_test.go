package cmd

import (
	"runtime"
	"testing"
)

func TestGetPlatformInfo(t *testing.T) {
	info := GetPlatformInfo()

	if info.OS == "" {
		t.Error("OS should not be empty")
	}
	if info.Arch == "" {
		t.Error("Arch should not be empty")
	}
	if info.ConfigDirSuffix == "" {
		t.Error("ConfigDirSuffix should not be empty")
	}
}

func TestGetConfigDirName(t *testing.T) {
	// 验证配置目录名称
	dirName := GetConfigDirName()
	if dirName != ".claude-switcher" {
		t.Errorf("expected config dir name to be .claude-switcher, got %s", dirName)
	}
}

func TestGetProfileDirName(t *testing.T) {
	// 验证配置子目录名称
	dirName := GetProfileDirName()
	if dirName != "profiles" {
		t.Errorf("expected profile dir name to be profiles, got %s", dirName)
	}
}

func TestGetActiveFileName(t *testing.T) {
	// 验证活动配置文件名
	name := GetActiveFileName()
	if name != "active" {
		t.Errorf("expected active file name to be active, got %s", name)
	}
}

func TestGetSettingsFileName(t *testing.T) {
	// 验证 settings 文件名
	name := GetSettingsFileName()
	if name != "settings.json" {
		t.Errorf("expected settings file name to be settings.json, got %s", name)
	}
}

func TestIsWindows(t *testing.T) {
	isWin := IsWindows()
	if runtime.GOOS == "windows" {
		if !isWin {
			t.Error("expected IsWindows to return true on Windows")
		}
	} else {
		if isWin {
			t.Error("expected IsWindows to return false on non-Windows")
		}
	}
}

func TestGetDefaultConfigDir(t *testing.T) {
	// 测试默认配置目录
	home := "/home/testuser"
	dir := GetDefaultConfigDir(home)

	expected := home + "/.claude-switcher"
	if dir != expected {
		t.Errorf("expected %s, got %s", expected, dir)
	}
}

func TestGetDefaultProfilesDir(t *testing.T) {
	// 测试默认配置目录
	home := "/home/testuser"
	configDir := home + "/.claude-switcher"
	dir := GetDefaultProfilesDir(configDir)

	expected := configDir + "/profiles"
	if dir != expected {
		t.Errorf("expected %s, got %s", expected, dir)
	}
}

func TestGetSettingsDir(t *testing.T) {
	// 测试 settings 目录
	home := "/home/testuser"
	dir := GetSettingsDir(home)

	expected := home + "/.claude"
	if dir != expected {
		t.Errorf("expected %s, got %s", expected, dir)
	}
}

func TestGetFilePermissions(t *testing.T) {
	// 测试文件权限
	perm := GetFilePermissions()
	if IsWindows() {
		// Windows 上权限设置应该宽松一些
		t.Logf("Windows permissions: %d", perm)
	} else {
		if perm != 0600 {
			t.Errorf("expected permissions 0600 on Unix, got %d", perm)
		}
	}
}

func TestNormalizePath(t *testing.T) {
	// 在非 Windows 系统上，路径应该保持不变
	// 在 Windows 系统上，反斜杠应该转换为正斜杠
	input := "path/to/file"
	expected := "path/to/file"

	result := NormalizePath(input)
	if result != expected {
		t.Errorf("NormalizePath(%s) = %s, want %s", input, result, expected)
	}
}
