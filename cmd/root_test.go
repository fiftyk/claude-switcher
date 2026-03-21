package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fiftyk/claude-switcher/internal/profile"
)

func TestGetSettingsFilePath(t *testing.T) {
	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, ".claude", "settings.json")

	if got := GetSettingsFilePath(); got != expected {
		t.Errorf("GetSettingsFilePath() = %v, want %v", got, expected)
	}
}

func TestBuildEnvVarsFromProfile(t *testing.T) {
	p := &profile.Profile{
		Name:      "test",
		AuthToken: "sk-test",
		BaseURL:   "https://api.example.com",
		HTTPProxy: "http://proxy:8080",
		Model:     "claude-3-5-sonnet",
		EnvVars:   map[string]string{},
	}

	envVars := BuildEnvVarsFromProfile(p)

	if envVars["ANTHROPIC_AUTH_TOKEN"] != "sk-test" {
		t.Errorf("ANTHROPIC_AUTH_TOKEN = %v, want %v", envVars["ANTHROPIC_AUTH_TOKEN"], "sk-test")
	}
	if envVars["ANTHROPIC_BASE_URL"] != "https://api.example.com" {
		t.Errorf("ANTHROPIC_BASE_URL = %v, want %v", envVars["ANTHROPIC_BASE_URL"], "https://api.example.com")
	}
	if envVars["http_proxy"] != "http://proxy:8080" {
		t.Errorf("http_proxy = %v, want %v", envVars["http_proxy"], "http://proxy:8080")
	}
	if envVars["https_proxy"] != "http://proxy:8080" {
		t.Errorf("https_proxy = %v, want %v", envVars["https_proxy"], "http://proxy:8080")
	}
	if envVars["ANTHROPIC_MODEL"] != "claude-3-5-sonnet" {
		t.Errorf("ANTHROPIC_MODEL = %v, want %v", envVars["ANTHROPIC_MODEL"], "claude-3-5-sonnet")
	}
}

func TestBuildEnvVarsFromProfileEmpty(t *testing.T) {
	p := &profile.Profile{
		Name: "empty",
	}

	envVars := BuildEnvVarsFromProfile(p)

	if len(envVars) != 0 {
		t.Errorf("envVars should be empty, got %v", envVars)
	}
}

// TestAppendEnvVars_MergeProfileOverrides tests that profile env vars override base env
func TestAppendEnvVars_MergeProfileOverrides(t *testing.T) {
	base := []string{"HOME=/home/user", "PATH=/usr/bin", "ANTHROPIC_AUTH_TOKEN=sk-old"}
	profileEnv := map[string]string{
		"ANTHROPIC_AUTH_TOKEN": "sk-new",
		"ANTHROPIC_BASE_URL":   "https://api.test.com",
	}

	merged := AppendEnvVars(base, profileEnv)

	envMap := make(map[string]string)
	for _, e := range merged {
		if idx := strings.Index(e, "="); idx >= 0 {
			envMap[e[:idx]] = e[idx+1:]
		}
	}

	if envMap["ANTHROPIC_AUTH_TOKEN"] != "sk-new" {
		t.Errorf("ANTHROPIC_AUTH_TOKEN = %q, want %q", envMap["ANTHROPIC_AUTH_TOKEN"], "sk-new")
	}
	if envMap["ANTHROPIC_BASE_URL"] != "https://api.test.com" {
		t.Errorf("ANTHROPIC_BASE_URL = %q, want %q", envMap["ANTHROPIC_BASE_URL"], "https://api.test.com")
	}
	if envMap["HOME"] != "/home/user" {
		t.Errorf("HOME = %q, want %q", envMap["HOME"], "/home/user")
	}
	if envMap["PATH"] != "/usr/bin" {
		t.Errorf("PATH = %q, want %q", envMap["PATH"], "/usr/bin")
	}
}

// TestAppendEnvVars_EmptyProfile tests with nil profile env vars
func TestAppendEnvVars_EmptyProfile(t *testing.T) {
	base := []string{"HOME=/home/user", "PATH=/usr/bin"}
	merged := AppendEnvVars(base, nil)
	if len(merged) != 2 {
		t.Errorf("len(merged) = %d, want 2", len(merged))
	}
}

// TestBuildEnvVarsFromProfile_AllFields tests every profile field maps correctly
func TestBuildEnvVarsFromProfile_AllFields(t *testing.T) {
	p := &profile.Profile{
		Name:       "all-fields",
		AuthToken:  "sk-all",
		BaseURL:    "https://all.example.com",
		HTTPProxy:  "http://allproxy:9090",
		HTTPSProxy: "http://allproxy-https:9091",
		Model:      "claude-3-opus-20240229",
		EnvVars: map[string]string{
			"EXTRA_KEY": "extra-value",
		},
	}

	got := BuildEnvVarsFromProfile(p)

	checks := map[string]string{
		"ANTHROPIC_AUTH_TOKEN": "sk-all",
		"ANTHROPIC_BASE_URL":   "https://all.example.com",
		"http_proxy":           "http://allproxy:9090",
		// HTTPSProxy is set but HTTPProxy takes precedence (both set to same value)
		"https_proxy":           "http://allproxy:9090",
		"ANTHROPIC_MODEL":      "claude-3-opus-20240229",
		"EXTRA_KEY":            "extra-value",
	}

	for k, want := range checks {
		if got[k] != want {
			t.Errorf("%s = %q, want %q", k, got[k], want)
		}
	}
}

// TestBuildEnvVarsFromProfile_HTTPSProxyOnly tests that when only HTTPSProxy is set,
// https_proxy follows HTTPProxy (HTTPSProxy is informational, not used by BuildEnvVarsFromProfile)
func TestBuildEnvVarsFromProfile_HTTPSProxyOnly(t *testing.T) {
	p := &profile.Profile{
		Name:       "https-only",
		HTTPSProxy: "http://https-only:8888",
	}

	got := BuildEnvVarsFromProfile(p)

	// BuildEnvVarsFromProfile only uses HTTPProxy, so neither proxy var is set
	if _, exists := got["https_proxy"]; exists {
		t.Error("https_proxy should not be set when only HTTPSProxy is configured (HTTPProxy is not set)")
	}
	if _, exists := got["http_proxy"]; exists {
		t.Error("http_proxy should not be set when only HTTPSProxy is configured")
	}
}

// TestRunClaudeWithProfile_Integration verifies RunClaudeWithEnv receives merged env vars
func TestRunClaudeWithProfile_Integration(t *testing.T) {
	profilesDir := t.TempDir()
	profileContent := `NAME="integration-test"
ANTHROPIC_AUTH_TOKEN="sk-integration"
ANTHROPIC_BASE_URL="https://integration.test"
http_proxy="http://proxy:3128"
`
	if err := os.WriteFile(filepath.Join(profilesDir, "integration-test.conf"), []byte(profileContent), 0600); err != nil {
		t.Fatal(err)
	}

	p, err := profile.LoadProfile(profilesDir, "integration-test")
	if err != nil {
		t.Fatal(err)
	}

	// Build env vars from profile
	profileEnv := BuildEnvVarsFromProfile(p)
	baseEnv := os.Environ()
	merged := AppendEnvVars(baseEnv, profileEnv)

	// Verify ANTHROPIC_AUTH_TOKEN is in merged env
	found := false
	for _, e := range merged {
		if e == "ANTHROPIC_AUTH_TOKEN=sk-integration" {
			found = true
			break
		}
	}
	if !found {
		t.Error("ANTHROPIC_AUTH_TOKEN=sk-integration not found in merged env")
	}
}

func TestGetExitIP(t *testing.T) {
	// 这个测试需要网络连接，简单验证函数执行
	ip, err := GetExitIP()
	// 可能失败（网络问题），但不应该是代码错误
	if err != nil && err.Error() != "curl failed" {
		t.Errorf("GetExitIP() unexpected error: %v", err)
	}
	// 如果成功，IP 应该是有效的格式
	if ip != "" && len(ip) < 7 {
		t.Errorf("GetExitIP() returned invalid IP: %v", ip)
	}
}

type mockProfile struct {
	Name       string
	AuthToken  string
	BaseURL    string
	HTTPProxy  string
	HTTPSProxy string
	Model      string
	EnvVars    map[string]string
}

func (m *mockProfile) GetEnvVars() map[string]string {
	return m.EnvVars
}

// 确保 mockProfile 实现了 ProfileInterface
var _ ProfileInterface = (*mockProfile)(nil)
