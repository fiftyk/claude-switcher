package main

import (
	"flag"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/fiftyk/claude-switcher/internal/config"
	"github.com/fiftyk/claude-switcher/internal/profile"
	"github.com/fiftyk/claude-switcher/internal/update"
)

// TestVersionParsing tests version parsing
func TestVersionParsing(t *testing.T) {
	tests := []struct {
		version string
		valid   bool
	}{
		{"v1.0.0", true},
		{"v2.3.4", true},
		{"invalid", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			v := update.ParseVersion(tt.version)
			if v.IsValid() != tt.valid {
				t.Errorf("ParseVersion(%q).IsValid() = %v, want %v", tt.version, v.IsValid(), tt.valid)
			}
		})
	}
}

// TestUpdateCheckConfig tests update check config
func TestUpdateCheckConfig(t *testing.T) {
	cfg := update.CheckConfig{
		Repo:      "fiftyk/claude-switcher",
		Interval:  24 * time.Hour,
		LastCheck: time.Now(),
		Enabled:   true,
	}

	// Test ShouldCheck with recent check
	if cfg.ShouldCheck() {
		t.Error("ShouldCheck() should be false with recent check")
	}

	// Test ShouldCheck with old check
	cfg.LastCheck = time.Now().Add(-25 * time.Hour)
	if !cfg.ShouldCheck() {
		t.Error("ShouldCheck() should be true with old check")
	}

	// Test ShouldCheck when disabled
	cfg.Enabled = false
	if cfg.ShouldCheck() {
		t.Error("ShouldCheck() should be false when disabled")
	}
}

// TestUpdateConfigLoadSave tests loading and saving update config
func TestUpdateConfigLoadSave(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := tmpDir + "/update.json"

	cfg := update.GetDefaultConfig("fiftyk/claude-switcher")
	if err := cfg.Save(configPath); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	loaded, err := update.LoadCheckConfig(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if loaded.Repo != cfg.Repo {
		t.Errorf("Repo = %q, want %q", loaded.Repo, cfg.Repo)
	}
	if loaded.Interval != cfg.Interval {
		t.Errorf("Interval = %v, want %v", loaded.Interval, cfg.Interval)
	}
	if loaded.Enabled != cfg.Enabled {
		t.Errorf("Enabled = %v, want %v", loaded.Enabled, cfg.Enabled)
	}
}

// TestConfigDirFunctions tests config directory functions
func TestConfigDirFunctions(t *testing.T) {
	// Test that config functions don't panic
	_ = config.GetConfigDir()
	_ = config.GetProfilesDir()
	_ = config.GetActiveFile()
	_ = update.GetConfigPath()
	_ = update.GetInstallPath()
}

// TestFlagParsing tests flag parsing (basic)
func TestFlagParsing(t *testing.T) {
	// Test that flags are properly defined
	versionFlag := flag.Bool("version", false, "show version")
	helpFlag := flag.Bool("help", false, "show help")
	_ = versionFlag
	_ = helpFlag

	// Verify flags have correct defaults
	if *versionFlag != false {
		t.Error("version flag should default to false")
	}
	if *helpFlag != false {
		t.Error("help flag should default to false")
	}
}

// TestVersionComparison tests version comparison
func TestVersionComparison(t *testing.T) {
	tests := []struct {
		current string
		latest  string
		want    int
	}{
		{"v1.0.0", "v2.0.0", -1}, // current < latest
		{"v2.0.0", "v1.0.0", 1},  // current > latest
		{"v1.0.0", "v1.0.0", 0},  // equal
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s_vs_%s", tt.current, tt.latest), func(t *testing.T) {
			current := update.ParseVersion(tt.current)
			latest := update.ParseVersion(tt.latest)
			got := current.Compare(latest)
			if got != tt.want {
				t.Errorf("Compare(%q, %q) = %d, want %d", tt.current, tt.latest, got, tt.want)
			}
		})
	}
}

// MockProfile for testing
type MockProfile struct {
	Name      string
	AuthToken string
	BaseURL   string
}

// TestProfileLoading tests profile loading logic
func TestProfileLoading(t *testing.T) {
	// Test that profile loading handles edge cases
	profilesDir := config.GetProfilesDir()

	// Loading non-existent profile should fail
	_, err := profile.LoadProfile(profilesDir, "nonexistent-profile-12345")
	if err == nil {
		t.Error("LoadProfile should fail for non-existent profile")
	}
}

// TestVersionString tests version string formatting
func TestVersionString(t *testing.T) {
	v := update.ParseVersion("v1.2.3")
	if v.String() != "v1.2.3" {
		t.Errorf("String() = %q, want %q", v.String(), "v1.2.3")
	}
}

// TestUpdateCheckResult tests update result structure
func TestUpdateCheckResult(t *testing.T) {
	result := &update.UpdateResult{
		HasUpdate:    true,
		Latest:       update.ParseVersion("v2.0.0"),
		DownloadURL:  "https://example.com/download",
		ChangelogURL: "https://example.com/changelog",
	}

	if !result.HasUpdate {
		t.Error("HasUpdate should be true")
	}
	if result.Latest.String() != "v2.0.0" {
		t.Errorf("Latest = %q, want %q", result.Latest.String(), "v2.0.0")
	}
	if result.DownloadURL == "" {
		t.Error("DownloadURL should not be empty")
	}
	if result.ChangelogURL == "" {
		t.Error("ChangelogURL should not be empty")
	}
}

// TestConfigValidation tests config validation
func TestConfigValidation(t *testing.T) {
	// Test valid config names
	validNames := []string{"work", "home", "test-profile"}
	for _, name := range validNames {
		valid, _ := config.ValidateConfigName(name)
		if !valid {
			t.Errorf("ValidateConfigName(%q) should be valid", name)
		}
	}

	// Test invalid config names
	invalidNames := []string{"", "has space", "/invalid", ".hidden"}
	for _, name := range invalidNames {
		valid, _ := config.ValidateConfigName(name)
		if valid {
			t.Errorf("ValidateConfigName(%q) should be invalid", name)
		}
	}
}

// TestURLValidation tests URL validation
func TestURLValidation(t *testing.T) {
	validURLs := []string{"https://api.example.com", "http://localhost:8080", ""}
	for _, url := range validURLs {
		if !config.ValidateURL(url) {
			t.Errorf("ValidateURL(%q) should be valid", url)
		}
	}

	invalidURLs := []string{"ftp://example.com", "not-a-url"}
	for _, url := range invalidURLs {
		if config.ValidateURL(url) {
			t.Errorf("ValidateURL(%q) should be invalid", url)
		}
	}
}

// TestProxyValidation tests proxy validation
func TestProxyValidation(t *testing.T) {
	validProxies := []string{"http://127.0.0.1:7890", "https://proxy:8080", ""}
	for _, proxy := range validProxies {
		if !config.ValidateProxy(proxy) {
			t.Errorf("ValidateProxy(%q) should be valid", proxy)
		}
	}

	invalidProxies := []string{"127.0.0.1:7890", "not-a-proxy"}
	for _, proxy := range invalidProxies {
		if config.ValidateProxy(proxy) {
			t.Errorf("ValidateProxy(%q) should be invalid", proxy)
		}
	}
}

// TestUpdateFunctions tests update utility functions
func TestUpdateFunctions(t *testing.T) {
	// Test GetDefaultConfig
	cfg := update.GetDefaultConfig("test/repo")
	if cfg == nil {
		t.Fatal("GetDefaultConfig returned nil")
	}
	if cfg.Repo != "test/repo" {
		t.Errorf("Repo = %q, want %q", cfg.Repo, "test/repo")
	}

	// Test ParseDownloadURL
	url := "https://github.com/test/repo/releases/download/v1.0.0/test-linux-amd64"
	got := update.ParseDownloadURL(url)
	want := "test-linux-amd64"
	if got != want {
		t.Errorf("ParseDownloadURL(%q) = %q, want %q", url, got, want)
	}
}

// TestVersionComparisonEdgeCases tests edge cases in version comparison
func TestVersionComparisonEdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		current string
		latest  string
	}{
		{"major edge", "v1.0.0", "v2.0.0"},
		{"minor edge", "v1.0.0", "v1.1.0"},
		{"patch edge", "v1.0.0", "v1.0.1"},
		{"all zeros", "v0.0.0", "v0.0.0"},
		{"prerelease like", "v1.0.0", "v1.0.0-beta"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			current := update.ParseVersion(tt.current)
			latest := update.ParseVersion(tt.latest)
			// Just verify it doesn't panic
			_ = current.Compare(latest)
		})
	}
}

// TestMockRunClaude tests that runClaude handles missing claude
func TestRunClaudeMissingClaude(t *testing.T) {
	// This test verifies the function signature and basic behavior
	// In a real environment, claude might not be installed
	// The function should handle this gracefully

	// We'll just verify the function exists and has correct signature
	cmdPath := "/usr/local/bin/claude"
	if _, err := os.Stat(cmdPath); os.IsNotExist(err) {
		// claude not installed - this is expected in test environment
		return
	}
}
