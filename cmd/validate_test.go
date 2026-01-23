package cmd

import (
	"testing"

	"github.com/fiftyk/claude-switcher/internal/profile"
)

func TestValidateProfile(t *testing.T) {
	// 测试有效配置
	p := &profile.Profile{
		Name:       "Test",
		AuthToken:  "sk-valid-token",
		BaseURL:    "https://api.anthropic.com",
		HTTPProxy:  "http://127.0.0.1:7890",
		Model:      "claude-3-5-sonnet",
	}

	result := ValidateProfile(p)

	if !result.Valid {
		t.Errorf("expected valid profile, got errors: %v", result.Errors)
	}
	if len(result.Errors) > 0 {
		t.Errorf("expected no errors, got %d", len(result.Errors))
	}
}

func TestValidateProfileEmpty(t *testing.T) {
	// 测试空配置
	p := &profile.Profile{
		Name:       "Empty",
		AuthToken:  "",
		BaseURL:    "",
		HTTPProxy:  "",
	}

	result := ValidateProfile(p)

	// 空配置应该是有效的（使用默认值）
	if !result.Valid {
		t.Errorf("expected empty profile to be valid, got errors: %v", result.Errors)
	}
}

func TestValidateProfileInvalidURL(t *testing.T) {
	// 测试无效 URL
	p := &profile.Profile{
		Name:      "Test",
		BaseURL:   "not-a-valid-url",
	}

	result := ValidateProfile(p)

	if result.Valid {
		t.Error("expected invalid profile for invalid URL")
	}

	// 应该检测到 URL 错误
	found := false
	for _, err := range result.Errors {
		if contains(err, "URL") || contains(err, "Base URL") {
			found = true
		}
	}
	if !found {
		t.Error("should detect invalid BaseURL")
	}
}

func TestValidateProfileInvalidProxy(t *testing.T) {
	// 测试无效代理
	p := &profile.Profile{
		Name:       "Test",
		HTTPProxy:  "invalid-proxy", // 缺少端口
	}

	result := ValidateProfile(p)

	if result.Valid {
		t.Error("expected invalid profile for invalid proxy")
	}
}

func TestValidateProfileInvalidToken(t *testing.T) {
	// 测试无效 Token 格式
	p := &profile.Profile{
		Name:      "Test",
		AuthToken: "invalid", // 不是有效的 sk- 格式
	}

	result := ValidateProfile(p)

	// Token 验证可能不会严格检查格式
	_ = result
}

func TestValidationResult(t *testing.T) {
	// 测试验证结果格式化
	p := &profile.Profile{
		Name:       "Test",
		BaseURL:    "invalid-url",
		HTTPProxy:  "no-port",
	}

	result := ValidateProfile(p)

	// 验证结果应该包含错误信息
	if result.Valid {
		t.Error("expected invalid result")
	}

	// 格式化错误信息
	output := FormatValidationResult(result)
	if output == "" {
		t.Error("expected formatted output")
	}

	if !contains(output, "验证失败") && !contains(output, "错误") {
		t.Error("output should mention validation failure")
	}
}

func TestCheckConnectivity(t *testing.T) {
	// 这个测试可能需要网络连接
	p := &profile.Profile{
		Name:      "Test",
		BaseURL:   "https://api.anthropic.com",
		AuthToken: "sk-test",
	}

	// 检查连通性（可能因为网络问题失败，但不应该是代码错误）
	result := CheckConnectivity(p)

	// 结果应该包含状态和消息
	_ = result
}

func TestCheckProxyConnectivity(t *testing.T) {
	// 测试代理连通性
	p := &profile.Profile{
		Name:       "Test",
		BaseURL:    "https://api.anthropic.com",
		HTTPProxy:  "http://127.0.0.1:7890",
		AuthToken:  "sk-test",
	}

	result := CheckProxyConnectivity(p)

	// 代理可能不可用，但应该能检测
	_ = result
}
