package settings

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Settings 表示 ~/.claude/settings.json 的结构
type Settings struct {
	Env                   map[string]string `json:"env,omitempty"`
	EnabledPlugins        map[string]bool   `json:"enabledPlugins,omitempty"`
	ClaudeSwitcherProfile string           `json:"_claudeSwitcherProfile,omitempty"`
}

// LoadSettings 从文件加载 settings.json
func LoadSettings(filePath string) (*Settings, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("settings.json 不存在: %w", err)
	}

	var s Settings
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("解析 settings.json 失败: %w", err)
	}

	// 初始化 map 以防 nil
	if s.Env == nil {
		s.Env = make(map[string]string)
	}
	if s.EnabledPlugins == nil {
		s.EnabledPlugins = make(map[string]bool)
	}

	return &s, nil
}

// SaveSettings 保存 settings.json
func SaveSettings(filePath string, s *Settings) error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化 settings.json 失败: %w", err)
	}

	// 确保目录存在
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0600); err != nil {
		return fmt.Errorf("写入 settings.json 失败: %w", err)
	}

	return nil
}

// SyncProfileToSettings 将 profile 的环境变量同步到 settings.json
func SyncProfileToSettings(filePath, profileName string, envVars map[string]string) error {
	var s *Settings

	// 如果文件存在，先加载以保留现有配置
	if _, err := os.Stat(filePath); err == nil {
		var err error
		s, err = LoadSettings(filePath)
		if err != nil {
			// 如果加载失败，创建新的
			s = &Settings{
				Env:          make(map[string]string),
				EnabledPlugins: make(map[string]bool),
			}
		}
	} else {
		s = &Settings{
			Env:          make(map[string]string),
			EnabledPlugins: make(map[string]bool),
		}
	}

	// 更新 env
	for k, v := range envVars {
		s.Env[k] = v
	}

	// 更新 profile 标记
	s.ClaudeSwitcherProfile = profileName

	// 保存
	return SaveSettings(filePath, s)
}