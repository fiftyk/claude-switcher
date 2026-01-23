package cmd

import (
	"testing"

	"github.com/fiftyk/claude-switcher/internal/profile"
)

func TestGetTemplateList(t *testing.T) {
	// 测试获取模板列表
	templates := GetTemplateList()
	if len(templates) == 0 {
		t.Error("expected at least one template")
	}

	// 验证每个模板都有名称和描述
	for _, tmpl := range templates {
		if tmpl.Name == "" {
			t.Error("template name should not be empty")
		}
		if tmpl.Description == "" {
			t.Error("template description should not be empty")
		}
	}
}

func TestApplyTemplate(t *testing.T) {
	templates := GetTemplateList()
	if len(templates) == 0 {
		t.Fatal("no templates available")
	}

	// 测试应用模板
	profile := ApplyTemplate(templates[0].Name, "test-profile")
	if profile == nil {
		t.Fatal("ApplyTemplate returned nil")
	}

	if profile.Name != "test-profile" {
		t.Errorf("expected profile name to be test-profile, got %s", profile.Name)
	}
}

func TestApplyTemplateNotFound(t *testing.T) {
	// 测试不存在的模板
	profile := ApplyTemplate("nonexistent", "test")
	if profile != nil {
		t.Error("expected nil for nonexistent template")
	}
}

func TestTemplatePresets(t *testing.T) {
	// 测试预设模板
	tests := []struct {
		templateName string
		checkFunc   func(*profile.Profile) bool
	}{
		{"default", func(p *profile.Profile) bool { return p.BaseURL == "" }},
		{"openai-compatible", func(p *profile.Profile) bool { return p.BaseURL != "" }},
		{"proxy", func(p *profile.Profile) bool { return p.HTTPProxy != "" }},
	}

	for _, tt := range tests {
		t.Run(tt.templateName, func(t *testing.T) {
			p := ApplyTemplate(tt.templateName, "test")
			if p == nil {
				t.Skip("template not found")
			}
			if !tt.checkFunc(p) {
				t.Errorf("template %s did not produce expected result", tt.templateName)
			}
		})
	}
}

func TestGetTemplateByName(t *testing.T) {
	// 测试通过名称获取模板
	tmpl := GetTemplateByName("default")
	if tmpl == nil {
		t.Error("expected to find default template")
	}
	if tmpl.Name != "default" {
		t.Errorf("expected template name to be default, got %s", tmpl.Name)
	}

	// 测试不存在的模板
	tmpl = GetTemplateByName("nonexistent")
	if tmpl != nil {
		t.Error("expected nil for nonexistent template")
	}
}
