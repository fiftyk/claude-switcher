package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/fiftyk/claude-switcher/internal/profile"
)

func TestDiffProfiles(t *testing.T) {
	// 配置文件1
	p1 := &profile.Profile{
		Name:       "Profile 1",
		AuthToken:  "sk-token-1",
		BaseURL:    "https://api.example.com",
		HTTPProxy:  "http://127.0.0.1:7890",
		Model:      "claude-3-5-sonnet",
		EnvVars:    map[string]string{"VAR1": "value1"},
	}

	// 配置文件2
	p2 := &profile.Profile{
		Name:       "Profile 2",
		AuthToken:  "sk-token-2",
		BaseURL:    "https://api.another.com",
		HTTPProxy:  "",
		Model:      "claude-3-haiku",
		EnvVars:    map[string]string{"VAR2": "value2"},
	}

	// 比较配置
	diff := DiffProfiles(p1, p2)

	// 验证差异检测
	if diff.HasDifferences {
		t.Log("Differences detected as expected")
	}

	// 验证各字段差异
	if len(diff.Differences) == 0 {
		t.Error("expected differences to be detected")
	}

	// 查找特定差异
	foundAuthToken := false
	foundBaseURL := false
	foundProxy := false
	for _, d := range diff.Differences {
		if d.Field == "AuthToken" {
			foundAuthToken = true
		}
		if d.Field == "BaseURL" {
			foundBaseURL = true
		}
		if d.Field == "HTTPProxy" {
			foundProxy = true
		}
	}

	if !foundAuthToken {
		t.Error("should detect AuthToken difference")
	}
	if !foundBaseURL {
		t.Error("should detect BaseURL difference")
	}
	if !foundProxy {
		t.Error("should detect HTTPProxy difference")
	}
}

func TestDiffProfilesIdentical(t *testing.T) {
	p1 := &profile.Profile{
		Name:       "Test Profile",
		AuthToken:  "sk-same-token",
		BaseURL:    "https://api.same.com",
	}

	p2 := &profile.Profile{
		Name:       "Test Profile",
		AuthToken:  "sk-same-token",
		BaseURL:    "https://api.same.com",
	}

	// 比较配置
	diff := DiffProfiles(p1, p2)

	// 相同配置应该没有差异
	if diff.HasDifferences {
		t.Error("identical profiles should not have differences")
	}
}

func TestFormatDiffOutput(t *testing.T) {
	diff := &ProfileDiff{
		HasDifferences: true,
		Profile1:       "Profile1",
		Profile2:       "Profile2",
		Differences: []FieldDiff{
			{Field: "AuthToken", Value1: "sk-token-1", Value2: "sk-token-2"},
			{Field: "BaseURL", Value1: "https://api.one.com", Value2: "https://api.two.com"},
		},
	}

	output := FormatDiffOutput(diff)

	// 验证输出包含必要内容
	if !contains(output, "Profile1") {
		t.Error("output should contain Profile1 name")
	}
	if !contains(output, "Profile2") {
		t.Error("output should contain Profile2 name")
	}
	if !contains(output, "AuthToken") {
		t.Error("output should contain AuthToken field")
	}
}

func TestLoadAndDiffProfiles(t *testing.T) {
	// 创建临时目录和两个测试配置
	tmpDir := t.TempDir()

	// 配置文件1
	content1 := `NAME="Config 1"
ANTHROPIC_AUTH_TOKEN="sk-config1"
ANTHROPIC_BASE_URL="https://api.config1.com"
`
	if err := os.WriteFile(filepath.Join(tmpDir, "config1.conf"), []byte(content1), 0600); err != nil {
		t.Fatal(err)
	}

	// 配置文件2
	content2 := `NAME="Config 2"
ANTHROPIC_AUTH_TOKEN="sk-config2"
ANTHROPIC_BASE_URL="https://api.config2.com"
`
	if err := os.WriteFile(filepath.Join(tmpDir, "config2.conf"), []byte(content2), 0600); err != nil {
		t.Fatal(err)
	}

	// 加载配置
	p1, err := profile.LoadProfile(tmpDir, "config1")
	if err != nil {
		t.Fatal(err)
	}
	p2, err := profile.LoadProfile(tmpDir, "config2")
	if err != nil {
		t.Fatal(err)
	}

	// 比较配置
	diff := DiffProfiles(p1, p2)

	// 应该有差异
	if !diff.HasDifferences {
		t.Error("expected differences between config1 and config2")
	}
}
