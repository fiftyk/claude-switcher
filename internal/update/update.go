package update

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// VersionInfo 表示版本信息
type VersionInfo struct {
	Major int
	Minor int
	Patch int
}

// ParseVersion 解析版本字符串
func ParseVersion(s string) VersionInfo {
	re := regexp.MustCompile(`^v(\d+)\.(\d+)\.(\d+)$`)
	matches := re.FindStringSubmatch(s)
	if matches == nil {
		return VersionInfo{}
	}
	major, _ := strconv.Atoi(matches[1])
	minor, _ := strconv.Atoi(matches[2])
	patch, _ := strconv.Atoi(matches[3])
	return VersionInfo{Major: major, Minor: minor, Patch: patch}
}

// IsValid 检查版本是否有效
func (v VersionInfo) IsValid() bool {
	return v.Major != 0 || v.Minor != 0 || v.Patch != 0
}

// String 返回版本字符串
func (v VersionInfo) String() string {
	return fmt.Sprintf("v%d.%d.%d", v.Major, v.Minor, v.Patch)
}

// Compare 比较两个版本
// 返回 -1 表示 current < latest, 0 表示相等, 1 表示 current > latest
func (v VersionInfo) Compare(other VersionInfo) int {
	if v.Major != other.Major {
		if v.Major < other.Major {
			return -1
		}
		return 1
	}
	if v.Minor != other.Minor {
		if v.Minor < other.Minor {
			return -1
		}
		return 1
	}
	if v.Patch != other.Patch {
		if v.Patch < other.Patch {
			return -1
		}
		return 1
	}
	return 0
}

// ReleaseInfo 表示 GitHub Release 信息
type ReleaseInfo struct {
	TagName string `json:"tag_name"`
}

// CheckConfig 表示自动更新检查配置
type CheckConfig struct {
	Repo      string
	Interval  time.Duration
	LastCheck time.Time
	Enabled   bool
}

// GetConfigPath 返回配置文件路径
func GetConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".claude-switcher", "update.json")
}

// Save 保存配置到文件
func (c *CheckConfig) Save(path string) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

// LoadCheckConfig 从文件加载配置
func LoadCheckConfig(path string) (*CheckConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg CheckConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// ShouldCheck 检查是否应该进行更新检查
func (c *CheckConfig) ShouldCheck() bool {
	if !c.Enabled {
		return false
	}
	if c.LastCheck.IsZero() {
		return true
	}
	return time.Since(c.LastCheck) >= c.Interval
}

// GetDefaultConfig 返回默认配置
func GetDefaultConfig(repo string) *CheckConfig {
	return &CheckConfig{
		Repo:     repo,
		Interval: 24 * time.Hour,
		Enabled:  true,
	}
}

// UpdateResult 表示更新检查结果
type UpdateResult struct {
	HasUpdate    bool
	Latest       VersionInfo
	DownloadURL  string
	ChangelogURL string
}

// CheckUpdate 检查是否有新版本
func CheckUpdate(repo string, current VersionInfo) (*UpdateResult, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("网络请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("GitHub API 返回错误: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	var release ReleaseInfo
	if err := json.Unmarshal(body, &release); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	latest := ParseVersion(release.TagName)
	hasUpdate := current.Compare(latest) < 0

	// 构建下载 URL
	os := runtime.GOOS
	arch := runtime.GOARCH
	binaryName := "claude-switcher"
	downloadURL := fmt.Sprintf("https://github.com/%s/releases/download/%s/%s-%s-%s",
		repo, release.TagName, binaryName, os, arch)

	return &UpdateResult{
		HasUpdate:    hasUpdate,
		Latest:       latest,
		DownloadURL:  downloadURL,
		ChangelogURL: fmt.Sprintf("https://github.com/%s/releases/tag/%s", repo, release.TagName),
	}, nil
}

// DownloadAndInstall 下载并安装新版本
func DownloadAndInstall(downloadURL, installPath string) error {
	// 下载文件
	resp, err := http.Get(downloadURL)
	if err != nil {
		return fmt.Errorf("下载失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("下载失败，状态码: %d", resp.StatusCode)
	}

	// 创建临时文件
	tmpFile, err := os.CreateTemp("", "claude-switcher-update-*")
	if err != nil {
		return fmt.Errorf("创建临时文件失败: %v", err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	// 保存到临时文件
	_, err = io.Copy(tmpFile, resp.Body)
	tmpFile.Close()
	if err != nil {
		return fmt.Errorf("保存文件失败: %v", err)
	}

	// 设置执行权限
	if err := os.Chmod(tmpPath, 0755); err != nil {
		return fmt.Errorf("设置权限失败: %v", err)
	}

	// 备份旧版本
	if _, err := os.Stat(installPath); err == nil {
		backupPath := installPath + ".backup." + time.Now().Format("20060102_150405")
		if err := os.Rename(installPath, backupPath); err != nil {
			return fmt.Errorf("备份失败: %v", err)
		}
	}

	// 移动新版本
	if err := os.Rename(tmpPath, installPath); err != nil {
		return fmt.Errorf("安装失败: %v", err)
	}

	return nil
}

// GetInstallPath 返回当前可执行文件的安装路径
func GetInstallPath() string {
	installDir := "/usr/local/bin"
	if runtime.GOOS == "darwin" {
		return filepath.Join(installDir, "claude-switcher")
	}
	return filepath.Join(installDir, "claude-switcher")
}

// IsInstallable 检查是否有权限安装
func IsInstallable(installPath string) bool {
	// 检查目录是否存在且可写
	dir := filepath.Dir(installPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return false
	}

	// 检查是否有写权限
	info, err := os.Stat(dir)
	if err == nil {
		return info.Mode().Perm()&0222 != 0
	}
	return false
}

// SelfUpdate 执行自更新
func SelfUpdate(repo string, current VersionInfo, installPath string) error {
	// 检查更新
	result, err := CheckUpdate(repo, current)
	if err != nil {
		return err
	}

	if !result.HasUpdate {
		return nil
	}

	// 下载并安装
	return DownloadAndInstall(result.DownloadURL, installPath)
}

// GetAutoCheckResult 获取自动检查结果（不打印信息）
func GetAutoCheckResult(repo string, current VersionInfo, lastCheck time.Time, interval time.Duration) (hasUpdate bool, latest VersionInfo, err error) {
	if lastCheck.IsZero() || time.Since(lastCheck) < interval {
		return false, VersionInfo{}, nil
	}

	result, err := CheckUpdate(repo, current)
	if err != nil {
		return false, VersionInfo{}, err
	}

	return result.HasUpdate, result.Latest, nil
}

// ParseDownloadURL 解析下载 URL 获取版本信息
func ParseDownloadURL(url string) string {
	parts := strings.Split(url, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ""
}

// Now 返回当前时间
func Now() time.Time {
	return time.Now()
}
