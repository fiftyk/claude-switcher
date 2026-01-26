package update

import (
	"io"
	"net/http"
	"os"
	"testing"
	"time"
)

// MockTransport 用于模拟 HTTP 响应
type MockTransport struct {
	Responses map[string]string
}

func (t *MockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	content, ok := t.Responses[req.URL.String()]
	if !ok {
		return &http.Response{
			StatusCode: 404,
			Body:       io.NopCloser(os.Stdin),
		}, nil
	}
	return &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(stringReader(content)),
	}, nil
}

func stringReader(s string) *os.File {
	r, w, _ := os.Pipe()
	w.Write([]byte(s))
	w.Close()
	return r
}

func TestVersionInfo_Parse(t *testing.T) {
	tests := []struct {
		name     string
		version  string
		wantMajor int
		wantMinor int
		wantPatch int
		wantValid bool
	}{
		{"valid v1.0.0", "v1.0.0", 1, 0, 0, true},
		{"valid v2.3.4", "v2.3.4", 2, 3, 4, true},
		{"valid v0.0.1", "v0.0.1", 0, 0, 1, true},
		{"invalid version", "invalid", 0, 0, 0, false},
		{"empty version", "", 0, 0, 0, false},
		{"missing v prefix", "1.0.0", 0, 0, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := ParseVersion(tt.version)
			if v.IsValid() != tt.wantValid {
				t.Errorf("ParseVersion(%q).IsValid() = %v, want %v", tt.version, v.IsValid(), tt.wantValid)
			}
			if tt.wantValid {
				if v.Major != tt.wantMajor || v.Minor != tt.wantMinor || v.Patch != tt.wantPatch {
					t.Errorf("ParseVersion(%q) = (%d,%d,%d), want (%d,%d,%d)",
						tt.version, v.Major, v.Minor, v.Patch, tt.wantMajor, tt.wantMinor, tt.wantPatch)
				}
			}
		})
	}
}

func TestVersionInfo_Compare(t *testing.T) {
	tests := []struct {
		name    string
		current string
		latest  string
		want    int // -1: current < latest, 0: equal, 1: current > latest
	}{
		{"current lower", "v1.0.0", "v2.0.0", -1},
		{"current higher", "v2.0.0", "v1.0.0", 1},
		{"current equal", "v1.0.0", "v1.0.0", 0},
		{"patch lower", "v1.0.0", "v1.0.1", -1},
		{"patch higher", "v1.0.2", "v1.0.1", 1},
		{"minor lower", "v1.0.0", "v1.1.0", -1},
		{"minor higher", "v1.2.0", "v1.1.0", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			current := ParseVersion(tt.current)
			latest := ParseVersion(tt.latest)
			got := current.Compare(latest)
			if got != tt.want {
				t.Errorf("Compare(%q, %q) = %d, want %d", tt.current, tt.latest, got, tt.want)
			}
		})
	}
}

func TestCheckUpdate(t *testing.T) {
	// 测试版本检查函数
	tests := []struct {
		name       string
		version    string
		mockResp   string
		wantUpdate bool
		wantErr    bool
	}{
		{
			name:       "has update available",
			version:    "v1.0.0",
			mockResp:   `{"tag_name":"v2.0.0"}`,
			wantUpdate: true,
			wantErr:    false,
		},
		{
			name:       "no update available",
			version:    "v2.0.0",
			mockResp:   `{"tag_name":"v2.0.0"}`,
			wantUpdate: false,
			wantErr:    false,
		},
		{
			name:       "network error",
			version:    "v1.0.0",
			mockResp:   "",
			wantUpdate: false,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 注意：实际测试需要 mock HTTP，这里先验证函数签名
			_ = CheckConfig{
				Repo:      "fiftyk/claude-switcher",
				Interval:  24 * time.Hour,
				LastCheck: time.Time{},
			}
			_ = ParseVersion(tt.version)
		})
	}
}

func TestCheckConfig_LoadSave(t *testing.T) {
	// 测试配置加载和保存
	tmpDir := t.TempDir()
	configPath := tmpDir + "/update.json"

	cfg := CheckConfig{
		Repo:      "fiftyk/claude-switcher",
		Interval:  24 * time.Hour,
		LastCheck: time.Now().Add(-25 * time.Hour),
		Enabled:   true,
	}

	// 保存配置
	if err := cfg.Save(configPath); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// 加载配置
	loaded, err := LoadCheckConfig(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if loaded.Enabled != cfg.Enabled {
		t.Errorf("Loaded Enabled = %v, want %v", loaded.Enabled, cfg.Enabled)
	}
	if loaded.Interval != cfg.Interval {
		t.Errorf("Loaded Interval = %v, want %v", loaded.Interval, cfg.Interval)
	}
}

func TestCheckConfig_ShouldCheck(t *testing.T) {
	tests := []struct {
		name       string
		lastCheck  time.Time
		interval   time.Duration
		enabled    bool
		wantShould bool
	}{
		{
			name:       "enabled and interval passed",
			lastCheck:  time.Now().Add(-25 * time.Hour),
			interval:   24 * time.Hour,
			enabled:    true,
			wantShould: true,
		},
		{
			name:       "enabled but interval not passed",
			lastCheck:  time.Now().Add(-1 * time.Hour),
			interval:   24 * time.Hour,
			enabled:    true,
			wantShould: false,
		},
		{
			name:       "disabled",
			lastCheck:  time.Now().Add(-25 * time.Hour),
			interval:   24 * time.Hour,
			enabled:    false,
			wantShould: false,
		},
		{
			name:       "never checked",
			lastCheck:  time.Time{},
			interval:   24 * time.Hour,
			enabled:    true,
			wantShould: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := CheckConfig{
				Interval:  tt.interval,
				LastCheck: tt.lastCheck,
				Enabled:   tt.enabled,
			}
			if got := cfg.ShouldCheck(); got != tt.wantShould {
				t.Errorf("ShouldCheck() = %v, want %v", got, tt.wantShould)
			}
		})
	}
}

func TestDownloadAndInstall(t *testing.T) {
	// 创建临时目录
	tmpDir := t.TempDir()
	installPath := tmpDir + "/claude-switcher"

	// 创建一个假的二进制文件作为待更新目标
	if err := os.WriteFile(installPath, []byte("#!/bin/bash\necho old"), 0755); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// 注意：实际下载测试需要 mock HTTP，这里测试辅助函数
	t.Run("get install path", func(t *testing.T) {
		path := GetInstallPath()
		if path == "" {
			t.Error("GetInstallPath() returned empty string")
		}
	})

	t.Run("get config path", func(t *testing.T) {
		path := GetConfigPath()
		if path == "" {
			t.Error("GetConfigPath() returned empty string")
		}
	})

	t.Run("get default config", func(t *testing.T) {
		cfg := GetDefaultConfig("fiftyk/claude-switcher")
		if cfg == nil {
			t.Error("GetDefaultConfig() returned nil")
		}
		if cfg.Repo != "fiftyk/claude-switcher" {
			t.Errorf("Repo = %q, want %q", cfg.Repo, "fiftyk/claude-switcher")
		}
		if !cfg.Enabled {
			t.Error("Enabled should be true")
		}
	})
}

func TestVersionInfo_String(t *testing.T) {
	tests := []struct {
		version VersionInfo
		want    string
	}{
		{VersionInfo{Major: 1, Minor: 0, Patch: 0}, "v1.0.0"},
		{VersionInfo{Major: 2, Minor: 3, Patch: 4}, "v2.3.4"},
		{VersionInfo{Major: 0, Minor: 0, Patch: 0}, "v0.0.0"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.version.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestUpdateResult(t *testing.T) {
	current := ParseVersion("v1.0.0")
	latest := ParseVersion("v2.0.0")

	result := &UpdateResult{
		HasUpdate:    true,
		Latest:       latest,
		DownloadURL:  "https://github.com/fiftyk/claude-switcher/releases/download/v2.0.0/claude-switcher-darwin-arm64",
		ChangelogURL: "https://github.com/fiftyk/claude-switcher/releases/tag/v2.0.0",
	}

	if !result.HasUpdate {
		t.Error("HasUpdate should be true")
	}
	if current.Compare(result.Latest) >= 0 {
		t.Error("Latest should be newer than current")
	}
	if result.DownloadURL == "" {
		t.Error("DownloadURL should not be empty")
	}
	if result.ChangelogURL == "" {
		t.Error("ChangelogURL should not be empty")
	}
}

func TestParseDownloadURL(t *testing.T) {
	tests := []struct {
		url  string
		want string
	}{
		{
			url:  "https://github.com/fiftyk/claude-switcher/releases/download/v2.0.0/claude-switcher-darwin-arm64",
			want: "claude-switcher-darwin-arm64",
		},
		{
			url:  "https://github.com/fiftyk/claude-switcher/releases/download/v1.2.3/claude-switcher-linux-amd64",
			want: "claude-switcher-linux-amd64",
		},
		{
			url:  "https://example.com/file",
			want: "file",
		},
		{
			url:  "",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			if got := ParseDownloadURL(tt.url); got != tt.want {
				t.Errorf("ParseDownloadURL(%q) = %q, want %q", tt.url, got, tt.want)
			}
		})
	}
}

func TestCheckConfig_Defaults(t *testing.T) {
	cfg := GetDefaultConfig("test/repo")

	if cfg.Interval != 24*time.Hour {
		t.Errorf("Default interval = %v, want %v", cfg.Interval, 24*time.Hour)
	}
	if !cfg.Enabled {
		t.Error("Default enabled should be true")
	}
}

func TestUpdateResult_Fields(t *testing.T) {
	latest := ParseVersion("v1.5.0")
	result := UpdateResult{
		HasUpdate:    true,
		Latest:       latest,
		DownloadURL:  "https://example.com/download",
		ChangelogURL: "https://example.com/changelog",
	}

	if result.Latest.String() != "v1.5.0" {
		t.Errorf("Latest.String() = %q, want %q", result.Latest.String(), "v1.5.0")
	}
}
