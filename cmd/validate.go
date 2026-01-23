package cmd

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/fiftyk/claude-switcher/internal/config"
	"github.com/fiftyk/claude-switcher/internal/profile"
)

// ValidationResult 验证结果
type ValidationResult struct {
	Valid   bool
	Errors  []string
	Warnings []string
}

// ValidateProfile 验证配置
func ValidateProfile(p *profile.Profile) *ValidationResult {
	result := &ValidationResult{
		Valid:   true,
		Errors:  []string{},
		Warnings: []string{},
	}

	// 验证配置名称
	if p.Name == "" {
		result.Valid = false
		result.Errors = append(result.Errors, "配置名称不能为空")
	}

	// 验证 Base URL
	if p.BaseURL != "" {
		if !config.ValidateURL(p.BaseURL) {
			result.Valid = false
			result.Errors = append(result.Errors, fmt.Sprintf("Base URL 格式无效: %s", p.BaseURL))
		}
	}

	// 验证代理
	if p.HTTPProxy != "" {
		if !config.ValidateProxy(p.HTTPProxy) {
			result.Valid = false
			result.Errors = append(result.Errors, fmt.Sprintf("HTTP Proxy 格式无效: %s", p.HTTPProxy))
		}
	}

	// 验证 Auth Token 格式（如果是提供的）
	if p.AuthToken != "" && !strings.HasPrefix(p.AuthToken, "sk-") {
		result.Warnings = append(result.Warnings, "Auth Token 可能不是有效的 Anthropic API Token")
	}

	// 检查 Model 是否为空（如果是提供的）
	if p.Model != "" {
		// 简单检查 model 名称格式
		if len(p.Model) < 5 {
			result.Warnings = append(result.Warnings, fmt.Sprintf("Model 名称可能无效: %s", p.Model))
		}
	}

	return result
}

// FormatValidationResult 格式化验证结果
func FormatValidationResult(result *ValidationResult) string {
	var sb strings.Builder

	if result.Valid {
		sb.WriteString("\n✓ 配置验证通过\n")
	} else {
		sb.WriteString("\n✗ 配置验证失败\n")
		for _, err := range result.Errors {
			sb.WriteString(fmt.Sprintf("  - %s\n", err))
		}
	}

	if len(result.Warnings) > 0 {
		sb.WriteString("\n⚠  警告:\n")
		for _, warn := range result.Warnings {
			sb.WriteString(fmt.Sprintf("  - %s\n", warn))
		}
	}

	return sb.String()
}

// ConnectivityResult 连通性检查结果
type ConnectivityResult struct {
	Reachable  bool
	Latency    time.Duration
	Message    string
	Error      error
}

// CheckConnectivity 检查 API 连通性
func CheckConnectivity(p *profile.Profile) *ConnectivityResult {
	result := &ConnectivityResult{
		Reachable: false,
		Message:   "未测试",
	}

	// 确定要检查的 URL
	checkURL := p.BaseURL
	if checkURL == "" {
		checkURL = "https://api.anthropic.com"
	}

	// 添加健康检查端点
	if !strings.HasSuffix(checkURL, "/") {
		checkURL += "/"
	}
	checkURL += "health"

	// 创建 HTTP 客户端
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// 设置代理
	if p.HTTPProxy != "" {
		// 注意：这里需要代理支持 HTTPS CONNECT
	}

	// 尝试请求
	start := time.Now()
	resp, err := client.Get(checkURL)
	latency := time.Since(start)

	if err != nil {
		result.Error = err
		result.Message = fmt.Sprintf("连接失败: %v", err)
		return result
	}
	defer resp.Body.Close()

	result.Latency = latency
	if resp.StatusCode == 200 || resp.StatusCode == 404 {
		result.Reachable = true
		result.Message = fmt.Sprintf("可达 (耗时 %v)", latency)
	} else {
		result.Message = fmt.Sprintf("返回状态码: %d", resp.StatusCode)
	}

	return result
}

// CheckProxyConnectivity 检查代理连通性
func CheckProxyConnectivity(p *profile.Profile) *ConnectivityResult {
	result := &ConnectivityResult{
		Reachable: false,
		Message:   "未测试",
	}

	if p.HTTPProxy == "" {
		result.Message = "未配置代理"
		return result
	}

	// 创建 HTTP 客户端，使用代理
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// 测试代理连通性（尝试访问百度）
	testURL := "https://www.baidu.com"

	start := time.Now()
	resp, err := client.Get(testURL)
	latency := time.Since(start)

	if err != nil {
		result.Error = err
		result.Message = fmt.Sprintf("代理连接失败: %v", err)
		return result
	}
	defer resp.Body.Close()

	result.Latency = latency
	if resp.StatusCode == 200 {
		result.Reachable = true
		result.Message = fmt.Sprintf("代理可达 (耗时 %v)", latency)
	} else {
		result.Message = fmt.Sprintf("返回状态码: %d", resp.StatusCode)
	}

	return result
}

// PrintValidationReport 打印完整验证报告
func PrintValidationReport(profilesDir, profileName string) error {
	p, err := profile.LoadProfile(profilesDir, profileName)
	if err != nil {
		return fmt.Errorf("无法加载配置: %w", err)
	}

	fmt.Printf("\n=== 配置验证: %s ===\n\n", profileName)

	// 格式验证
	fmt.Println("格式验证:")
	fmt.Println(strings.Repeat("-", 40))
	validation := ValidateProfile(p)
	fmt.Print(FormatValidationResult(validation))

	// 连通性检查
	fmt.Println("\nAPI 连通性:")
	fmt.Println(strings.Repeat("-", 40))
	conn := CheckConnectivity(p)
	if conn.Reachable {
		fmt.Printf("  ✓ API 可达 %s\n", conn.Message)
	} else {
		fmt.Printf("  ✗ API 不可达 %s\n", conn.Message)
	}

	// 代理检查
	if p.HTTPProxy != "" {
		fmt.Println("\n代理连通性:")
		fmt.Println(strings.Repeat("-", 40))
		proxyConn := CheckProxyConnectivity(p)
		if proxyConn.Reachable {
			fmt.Printf("  ✓ 代理可达 %s\n", proxyConn.Message)
		} else {
			fmt.Printf("  ✗ 代理不可达 %s\n", proxyConn.Message)
		}
	}

	fmt.Println()
	return nil
}

// PrintValidateHelp 打印验证帮助信息
func PrintValidateHelp() {
	fmt.Println("\n配置验证用法:")
	fmt.Println("  claude-switcher --validate <配置名>   验证配置格式")
	fmt.Println("  claude-switcher --test <配置名>       验证配置并测试连通性")
	fmt.Println()
}
