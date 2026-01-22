package profile

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Profile 表示一个 Claude 配置
type Profile struct {
	Name      string
	AuthToken string
	BaseURL   string
	HTTPProxy string
	HTTPSProxy string
	Model     string
	EnvVars   map[string]string
}

// LoadProfile 从文件加载配置
func LoadProfile(profilesDir, name string) (*Profile, error) {
	filePath := filepath.Join(profilesDir, name+".conf")
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("配置文件不存在: %s", name)
	}
	defer file.Close()

	p := &Profile{
		EnvVars: make(map[string]string),
	}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// 跳过空行和注释行
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// 解析 VAR=value 格式
		kv := parseLine(line)
		if kv == nil {
			continue
		}

		key, value := kv[0], kv[1]

		// 根据键名设置对应字段
		switch key {
		case "NAME":
			p.Name = value
		case "ANTHROPIC_AUTH_TOKEN":
			p.AuthToken = value
		case "ANTHROPIC_BASE_URL":
			p.BaseURL = value
		case "http_proxy":
			p.HTTPProxy = value
			// 如果没有设置 https_proxy，使用相同的值
			if p.HTTPSProxy == "" {
				p.HTTPSProxy = value
			}
		case "https_proxy":
			p.HTTPSProxy = value
		case "ANTHROPIC_MODEL":
			p.Model = value
		default:
			// 其他变量放入 EnvVars
			if !strings.HasPrefix(key, "_") {
				p.EnvVars[key] = value
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return p, nil
}

func parseLine(line string) []string {
	// 匹配 VAR=value 格式
	re := regexp.MustCompile(`^([A-Za-z_][A-Za-z0-9_]*)=(.*)$`)
	matches := re.FindStringSubmatch(line)
	if matches == nil {
		return nil
	}

	key := matches[1]
	value := strings.Trim(matches[2], `"'`)

	return []string{key, value}
}

// ListProfiles 列出所有配置名称
func ListProfiles(profilesDir string) ([]string, error) {
	var names []string
	entries, err := os.ReadDir(profilesDir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasSuffix(name, ".conf") {
			names = append(names, strings.TrimSuffix(name, ".conf"))
		}
	}

	return names, nil
}

// DeleteProfile 删除配置
func DeleteProfile(profilesDir, name string) error {
	filePath := filepath.Join(profilesDir, name+".conf")
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("配置文件不存在: %s", name)
	}
	return os.Remove(filePath)
}

// CopyProfile 复制配置
func CopyProfile(profilesDir, srcName, dstName string) error {
	srcFile := filepath.Join(profilesDir, srcName+".conf")
	dstFile := filepath.Join(profilesDir, dstName+".conf")

	if _, err := os.Stat(srcFile); os.IsNotExist(err) {
		return fmt.Errorf("源配置不存在: %s", srcName)
	}

	content, err := os.ReadFile(srcFile)
	if err != nil {
		return err
	}

	// 更新 NAME 字段
	re := regexp.MustCompile(`(?m)^NAME=.*$`)
	content = re.ReplaceAllFunc(content, func(match []byte) []byte {
		return []byte(fmt.Sprintf(`NAME="%s"`, dstName))
	})

	return os.WriteFile(dstFile, content, 0600)
}

// RenameProfile 重命名配置
func RenameProfile(profilesDir, oldName, newName string) error {
	srcFile := filepath.Join(profilesDir, oldName+".conf")

	if _, err := os.Stat(srcFile); os.IsNotExist(err) {
		return fmt.Errorf("配置不存在: %s", oldName)
	}

	// 先复制
	if err := CopyProfile(profilesDir, oldName, newName); err != nil {
		return err
	}

	// 删除原文件
	return os.Remove(srcFile)
}
