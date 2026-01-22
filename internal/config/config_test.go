package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetConfigDir(t *testing.T) {
	// 测试 GetConfigDir 返回正确的目录
	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, ".claude-switcher")

	if got := GetConfigDir(); got != expected {
		t.Errorf("GetConfigDir() = %v, want %v", got, expected)
	}
}

func TestGetProfilesDir(t *testing.T) {
	// 测试 GetProfilesDir 返回正确的目录
	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, ".claude-switcher", "profiles")

	if got := GetProfilesDir(); got != expected {
		t.Errorf("GetProfilesDir() = %v, want %v", got, expected)
	}
}

func TestGetActiveFile(t *testing.T) {
	// 测试 GetActiveFile 返回正确的路径
	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, ".claude-switcher", "active")

	if got := GetActiveFile(); got != expected {
		t.Errorf("GetActiveFile() = %v, want %v", got, expected)
	}
}

func TestEnsureConfigDir(t *testing.T) {
	// 测试 EnsureConfigDir 创建目录
	tmpDir := t.TempDir()
	originalConfigDir := ConfigDir
	ConfigDir = filepath.Join(tmpDir, ".claude-switcher")
	defer func() { ConfigDir = originalConfigDir }()

	if err := EnsureConfigDir(); err != nil {
		t.Errorf("EnsureConfigDir() error = %v", err)
	}

	if _, err := os.Stat(ConfigDir); os.IsNotExist(err) {
		t.Errorf("EnsureConfigDir() did not create directory")
	}
}

func TestValidateConfigName(t *testing.T) {
	tests := []struct {
		name    string
		want    bool
		wantErr string
	}{
		{"valid-name", true, ""},
		{"valid_name_123", true, ""},
		{"", false, "配置名称不能为空"},
		{"name with space", false, "配置名称不能包含空格"},
		{"name/", false, "配置名称不能包含特殊字符"},
		{"name\\", false, "配置名称不能包含特殊字符"},
		{"name..", false, "配置名称不能包含特殊字符"},
		{"name~", false, "配置名称不能包含特殊字符"},
		{"name$", false, "配置名称不能包含特殊字符"},
		{".hidden", false, "配置名称不能以点开头或结尾"},
		{"hidden.", false, "配置名称不能以点开头或结尾"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ValidateConfigName(tt.name)
			if got != tt.want {
				t.Errorf("ValidateConfigName(%v) = %v, want %v", tt.name, got, tt.want)
			}
			if err != nil && err.Error() != tt.wantErr {
				t.Errorf("ValidateConfigName(%v) error = %v, wantErr %v", tt.name, err, tt.wantErr)
			}
		})
	}
}

func TestValidateURL(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want bool
	}{
		{"empty", "", true},
		{"http", "http://example.com", true},
		{"https", "https://api.anthropic.com", true},
		{"https with path", "https://api.example.com/v1", true},
		{"invalid", "ftp://example.com", false},
		{"no protocol", "example.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidateURL(tt.url); got != tt.want {
				t.Errorf("ValidateURL(%v) = %v, want %v", tt.url, got, tt.want)
			}
		})
	}
}

func TestValidateProxy(t *testing.T) {
	tests := []struct {
		name  string
		proxy string
		want  bool
	}{
		{"empty", "", true},
		{"http", "http://127.0.0.1:7890", true},
		{"https", "https://proxy.example.com:8080", true},
		{"no port", "http://127.0.0.1", false},
		{"no protocol", "127.0.0.1:7890", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidateProxy(tt.proxy); got != tt.want {
				t.Errorf("ValidateProxy(%v) = %v, want %v", tt.proxy, got, tt.want)
			}
		})
	}
}
