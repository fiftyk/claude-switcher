package settings

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Settings 表示 ~/.claude/settings.json 的结构
type Settings struct {
	Env                   map[string]string          `json:"env,omitempty"`
	EnabledPlugins        map[string]bool             `json:"enabledPlugins,omitempty"`
	ClaudeSwitcherProfile string                      `json:"_claudeSwitcherProfile,omitempty"`
	Extra                 map[string]json.RawMessage  `json:"-"` // 保留未知字段
}

// UnmarshalJSON 自定义反序列化，保留未知字段
func (s *Settings) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	if v, ok := raw["env"]; ok {
		json.Unmarshal(v, &s.Env)
		delete(raw, "env")
	}
	if v, ok := raw["enabledPlugins"]; ok {
		json.Unmarshal(v, &s.EnabledPlugins)
		delete(raw, "enabledPlugins")
	}
	if v, ok := raw["_claudeSwitcherProfile"]; ok {
		json.Unmarshal(v, &s.ClaudeSwitcherProfile)
		delete(raw, "_claudeSwitcherProfile")
	}

	if s.Env == nil {
		s.Env = make(map[string]string)
	}
	if s.EnabledPlugins == nil {
		s.EnabledPlugins = make(map[string]bool)
	}

	s.Extra = raw
	return nil
}

// MarshalJSON 自定义序列化，保留未知字段
func (s Settings) MarshalJSON() ([]byte, error) {
	raw := make(map[string]json.RawMessage)

	// 先放未知字段
	for k, v := range s.Extra {
		raw[k] = v
	}

	// 再放已知字段
	if len(s.Env) > 0 {
		v, _ := json.Marshal(s.Env)
		raw["env"] = v
	}
	if len(s.EnabledPlugins) > 0 {
		v, _ := json.Marshal(s.EnabledPlugins)
		raw["enabledPlugins"] = v
	}
	if s.ClaudeSwitcherProfile != "" {
		v, _ := json.Marshal(s.ClaudeSwitcherProfile)
		raw["_claudeSwitcherProfile"] = v
	}

	return json.Marshal(raw)
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

	// 清空旧的 env，再写入新 profile 的环境变量
	s.Env = make(map[string]string)
	for k, v := range envVars {
		s.Env[k] = v
	}

	// 更新 profile 标记
	s.ClaudeSwitcherProfile = profileName

	// 保存
	return SaveSettings(filePath, s)
}

// ClearProfileEnvVars 清除 settings.json 中与指定 profile 相关的环境变量
// 这样可以让运行时 profile 的环境变量优先于 settings.json
func ClearProfileEnvVars(filePath string, profileName string) error {
	// 如果文件不存在，直接返回
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil
	}

	s, err := LoadSettings(filePath)
	if err != nil {
		return err
	}

	// 需要清除的环境变量列表
	keysToClear := []string{
		"ANTHROPIC_AUTH_TOKEN",
		"ANTHROPIC_BASE_URL",
		"http_proxy",
		"https_proxy",
		"ANTHROPIC_MODEL",
	}

	// 清除这些键
	for _, key := range keysToClear {
		delete(s.Env, key)
	}

	// 清除 profile 标记
	s.ClaudeSwitcherProfile = ""

	return SaveSettings(filePath, s)
}