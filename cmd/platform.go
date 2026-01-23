package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// PlatformInfo 平台信息
type PlatformInfo struct {
	OS              string
	Arch            string
	ConfigDirSuffix string
}

// GetPlatformInfo 获取当前平台信息
func GetPlatformInfo() *PlatformInfo {
	return &PlatformInfo{
		OS:              runtime.GOOS,
		Arch:            runtime.GOARCH,
		ConfigDirSuffix: ".claude-switcher",
	}
}

// GetConfigDirName 获取配置目录名称
func GetConfigDirName() string {
	return ".claude-switcher"
}

// GetProfileDirName 获取配置子目录名称
func GetProfileDirName() string {
	return "profiles"
}

// GetActiveFileName 获取活动配置文件名
func GetActiveFileName() string {
	return "active"
}

// GetSettingsFileName 获取 settings.json 文件名
func GetSettingsFileName() string {
	return "settings.json"
}

// IsWindows 判断是否为 Windows 平台
func IsWindows() bool {
	return runtime.GOOS == "windows"
}

// GetDefaultConfigDir 获取默认配置目录
func GetDefaultConfigDir(home string) string {
	return filepath.Join(home, GetConfigDirName())
}

// GetDefaultProfilesDir 获取默认 profiles 目录
func GetDefaultProfilesDir(configDir string) string {
	return filepath.Join(configDir, GetProfileDirName())
}

// GetSettingsDir 获取 settings.json 所在目录
func GetSettingsDir(home string) string {
	return filepath.Join(home, ".claude")
}

// GetFilePermissions 获取文件权限
// Windows 上使用较宽松的权限，Unix 上使用严格的 0600
func GetFilePermissions() os.FileMode {
	if IsWindows() {
		return 0644 // Windows 上不支持 Unix 权限位
	}
	return 0600
}

// NormalizePath 规范化路径（Windows 路径分隔符转换）
func NormalizePath(path string) string {
	if IsWindows() {
		// Windows 使用反斜杠，但内部统一使用正斜杠
		return strings.ReplaceAll(path, "\\", "/")
	}
	return path
}

// GetSettingsFilePath 获取 settings.json 完整路径
func GetSettingsFilePath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(GetSettingsDir(home), GetSettingsFileName())
}

// EnsureDirExists 确保目录存在
func EnsureDirExists(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return os.MkdirAll(path, 0755)
	}
	return nil
}

// IsValidPathChar 检查路径字符是否有效
func IsValidPathChar(c byte) bool {
	if c >= 'a' && c <= 'z' {
		return true
	}
	if c >= 'A' && c <= 'Z' {
		return true
	}
	if c >= '0' && c <= '9' {
		return true
	}
	if c == '-' || c == '_' || c == '.' {
		return true
	}
	return false
}

// SanitizePathComponent 清理路径组件
func SanitizePathComponent(name string) string {
	var result []byte
	for i := 0; i < len(name); i++ {
		if IsValidPathChar(name[i]) {
			result = append(result, name[i])
		} else {
			result = append(result, '_')
		}
	}
	return string(result)
}

// PrintPlatformInfo 打印平台信息
func PrintPlatformInfo() {
	info := GetPlatformInfo()
	fmt.Printf("\n平台信息:\n")
	fmt.Printf("  操作系统: %s\n", info.OS)
	fmt.Printf("  架构: %s\n", info.Arch)
	fmt.Printf("  配置目录: ~/%s\n", GetConfigDirName())
	fmt.Printf("  配置目录位置: %s\n", GetDefaultConfigDir("/home/testuser"))
	fmt.Println()
}

// PrintPlatformHelp 打印平台相关帮助
func PrintPlatformHelp() {
	fmt.Println("\n跨平台支持:")
	fmt.Println("  claude-switcher 支持 macOS、Linux 和 Windows")
	fmt.Println()
	fmt.Println("配置文件位置:")
	fmt.Println("  macOS/Linux: ~/.claude-switcher/")
	winPath := "%USERPROFILE%" + string(os.PathSeparator) + ".claude-switcher"
	fmt.Println("  Windows:", winPath)
	fmt.Println()
	fmt.Println("Windows 注意事项:")
	fmt.Println("  - 使用 cmd.exe 时: claude-switcher <配置名>")
	fmt.Println("  - 使用 PowerShell 时: claude-switcher <配置名>")
	fmt.Println("  - 文件权限设置较宽松")
	fmt.Println()
}
