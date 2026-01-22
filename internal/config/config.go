package config

import (
	"os"
	"path/filepath"
	"regexp"
)

// ConfigDir 是配置目录路径
var ConfigDir string

// GetConfigDir 返回配置目录路径
func GetConfigDir() string {
	if ConfigDir == "" {
		home, _ := os.UserHomeDir()
		ConfigDir = filepath.Join(home, ".claude-switcher")
	}
	return ConfigDir
}

// GetProfilesDir 返回 profiles 目录路径
func GetProfilesDir() string {
	return filepath.Join(GetConfigDir(), "profiles")
}

// GetActiveFile 返回活动配置记录文件路径
func GetActiveFile() string {
	return filepath.Join(GetConfigDir(), "active")
}

// EnsureConfigDir 确保配置目录存在
func EnsureConfigDir() error {
	dir := GetConfigDir()
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return os.MkdirAll(dir, 0700)
	}
	return nil
}

// ValidateConfigName 验证配置名称
func ValidateConfigName(name string) (bool, error) {
	if name == "" {
		return false, nil
	}

	if len(name) > 50 {
		return false, nil
	}

	// 检查特殊字符
	specialChars := regexp.MustCompile(`[/.\\~*$]`)
	if specialChars.MatchString(name) {
		return false, nil
	}

	// 检查首尾点
	if name[0] == '.' || name[len(name)-1] == '.' {
		return false, nil
	}

	// 检查空格
	if containsSpace(name) {
		return false, nil
	}

	return true, nil
}

func containsSpace(s string) bool {
	for _, c := range s {
		if c == ' ' {
			return true
		}
	}
	return false
}

// ValidateURL 验证 URL 格式
func ValidateURL(url string) bool {
	if url == "" {
		return true
	}
	matched, _ := regexp.MatchString(`^https?://[a-zA-Z0-9.-]+(:[0-9]+)?([/][^[:space:]]*)?$`, url)
	return matched
}

// ValidateProxy 验证代理格式
func ValidateProxy(proxy string) bool {
	if proxy == "" {
		return true
	}
	matched, _ := regexp.MatchString(`^https?://[a-zA-Z0-9.-]+:[0-9]+$`, proxy)
	return matched
}
