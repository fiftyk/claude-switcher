package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/fiftyk/claude-switcher/cmd"
	"github.com/fiftyk/claude-switcher/internal/config"
	"github.com/fiftyk/claude-switcher/internal/profile"
	"github.com/fiftyk/claude-switcher/internal/settings"
)

// 版本信息 (由 Go Releaser 注入)
var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

// 版本信息
const appName = "Claude Switcher"

func main() {
	// 解析参数
	configName := flag.String("config", "", "指定配置名称")
	syncSettings := flag.Bool("sync", false, "同步到 settings.json")
	listFlag := flag.Bool("list", false, "列出所有配置")
	testFlag := flag.String("test", "", "测试配置")
	renameFlag := flag.String("rename", "", "重命名配置")
	copyFlag := flag.String("copy", "", "复制配置")
	helpFlag := flag.Bool("help", false, "显示帮助")
	versionFlag := flag.Bool("version", false, "显示版本信息")
	flag.Parse()

	// 显示版本
	if *versionFlag {
		showVersion()
		return
	}

	// 显示帮助
	if *helpFlag || len(os.Args) == 1 {
		showHelp()
		return
	}

	// 初始化配置目录
	if err := config.EnsureConfigDir(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	profilesDir := config.GetProfilesDir()

	// 处理透传参数（-- 之后的部分）
	var forwardArgs []string
	var configNameFromArgs string
	args := os.Args[1:]

	// 检测是否有 -- 分隔符
	if idx := indexOf(args, "--"); idx >= 0 {
		configNameFromArgs = args[0]
		forwardArgs = args[idx+1:]
		args = args[:idx]
	} else {
		// 检查是否有其他参数形式
		if len(args) >= 2 {
			if args[0] == "-c" || args[0] == "--config" {
				configNameFromArgs = args[1]
				args = args[2:]
			} else if !strings.HasPrefix(args[1], "-") && args[1] != "--sync" {
				// 可能是 <name> <arg> 形式
				configNameFromArgs = args[0]
				forwardArgs = args[1:]
				args = []string{}
			} else if args[1] == "--sync" {
				configNameFromArgs = args[0]
				*syncSettings = true
				args = args[2:]
			}
		} else if len(args) == 1 {
			configNameFromArgs = args[0]
			args = []string{}
		}
	}

	// 使用 flag 解析的值作为后备
	if *configName != "" {
		configNameFromArgs = *configName
	}

	// 处理命令
	switch {
	case *listFlag:
		listProfiles(profilesDir)
		return
	case *testFlag != "":
		testConfig(profilesDir, *testFlag)
		return
	case *renameFlag != "":
		parts := strings.Split(*renameFlag, " ")
		if len(parts) != 2 {
			fmt.Fprintln(os.Stderr, "Error: 用法 --rename <旧名称> <新名称>")
			os.Exit(1)
		}
		renameProfile(profilesDir, parts[0], parts[1])
		return
	case *copyFlag != "":
		parts := strings.Split(*copyFlag, " ")
		if len(parts) != 2 {
			fmt.Fprintln(os.Stderr, "Error: 用法 --copy <源名称> <目标名称>")
			os.Exit(1)
		}
		copyProfile(profilesDir, parts[0], parts[1])
		return
	case configNameFromArgs != "":
		// 加载配置
		p, err := profile.LoadProfile(profilesDir, configNameFromArgs)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			listProfiles(profilesDir)
			os.Exit(1)
		}

		// 验证配置名称
		if valid, _ := config.ValidateConfigName(configNameFromArgs); !valid {
			fmt.Fprintln(os.Stderr, "Error: 配置名称格式不正确")
			os.Exit(1)
		}

		// 如果需要同步到 settings.json
		if *syncSettings {
			if err := syncToSettings(configNameFromArgs, p); err != nil {
				fmt.Fprintf(os.Stderr, "Error: 同步到 settings.json 失败: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("✓ 已同步到 settings.json")
		}

		// 设置活动配置
		if err := setActiveProfile(configNameFromArgs); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("使用配置: %s\n", configNameFromArgs)
		runClaude(forwardArgs...)
		return
	default:
		showHelp()
	}
}

func indexOf(args []string, target string) int {
	for i, arg := range args {
		if arg == target {
			return i
		}
	}
	return -1
}

func showHelp() {
	fmt.Printf(`%s v%s - 使用帮助

用法:
  claude-switcher                    启动交互式配置选择
  claude-switcher <配置名称> [-- <参数...>]  使用指定配置启动，可透传参数
  claude-switcher <配置名称> --sync         切换配置并同步到 settings.json
  claude-switcher --config <名称> [-- <参数...>] 使用指定配置启动，可透传参数
  claude-switcher --list             列出所有可用配置
  claude-switcher --test <名称>      测试配置有效性
  claude-switcher --rename <旧> <新> 重命名配置
  claude-switcher --copy <源> <目标>  复制配置
  claude-switcher --version          显示版本信息
  claude-switcher --help             显示此帮助信息

说明:
  • 配置文件位于: ~/.claude-switcher/profiles/
  • 无参数运行时进入交互式菜单
  • 使用 --sync 参数可将配置同步到 ~/.claude/settings.json

`, appName, version)
}

func showVersion() {
	fmt.Printf("%s version %s (commit: %s, date: %s)\n", appName, version, commit, date)
}

func listProfiles(profilesDir string) {
	names, err := profile.ListProfiles(profilesDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return
	}

	fmt.Println("可用配置:")
	for _, name := range names {
		p, err := profile.LoadProfile(profilesDir, name)
		if err != nil {
			continue
		}
		displayName := name
		if p.Name != "" {
			displayName = p.Name
		}
		fmt.Printf("  %s - %s\n", name, displayName)
	}
}

func testConfig(profilesDir, name string) {
	_, err := profile.LoadProfile(profilesDir, name)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✓ 配置有效")
}

func renameProfile(profilesDir, oldName, newName string) {
	if valid, _ := config.ValidateConfigName(oldName); !valid {
		fmt.Fprintln(os.Stderr, "Error: 旧配置名称格式不正确")
		os.Exit(1)
	}
	if valid, _ := config.ValidateConfigName(newName); !valid {
		fmt.Fprintln(os.Stderr, "Error: 新配置名称格式不正确")
		os.Exit(1)
	}

	if err := profile.RenameProfile(profilesDir, oldName, newName); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✓ 已重命名: %s -> %s\n", oldName, newName)
}

func copyProfile(profilesDir, srcName, dstName string) {
	if valid, _ := config.ValidateConfigName(srcName); !valid {
		fmt.Fprintln(os.Stderr, "Error: 源配置名称格式不正确")
		os.Exit(1)
	}
	if valid, _ := config.ValidateConfigName(dstName); !valid {
		fmt.Fprintln(os.Stderr, "Error: 目标配置名称格式不正确")
		os.Exit(1)
	}

	if err := profile.CopyProfile(profilesDir, srcName, dstName); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✓ 已复制: %s -> %s\n", srcName, dstName)
}

func syncToSettings(profileName string, p *profile.Profile) error {
	settingsPath := cmd.GetSettingsFilePath()

	// 构建环境变量
	envVars := make(map[string]string)
	if p.AuthToken != "" {
		envVars["ANTHROPIC_AUTH_TOKEN"] = p.AuthToken
	}
	if p.BaseURL != "" {
		envVars["ANTHROPIC_BASE_URL"] = p.BaseURL
	}
	if p.HTTPProxy != "" {
		envVars["http_proxy"] = p.HTTPProxy
		envVars["https_proxy"] = p.HTTPProxy
	}
	for k, v := range p.EnvVars {
		envVars[k] = v
	}

	return settings.SyncProfileToSettings(settingsPath, profileName, envVars)
}

func setActiveProfile(name string) error {
	activeFile := config.GetActiveFile()
	return os.WriteFile(activeFile, []byte(name), 0600)
}

func runClaude(args ...string) {
	// 检查 claude 是否安装
	if _, err := os.Stat("/usr/local/bin/claude"); err != nil {
		// 尝试在 PATH 中查找
		if _, err := exec.LookPath("claude"); err != nil {
			fmt.Fprintln(os.Stderr, "Error: Claude CLI 未安装")
			os.Exit(1)
		}
	}

	// 构建命令
	cmdArgs := append([]string{}, args...)
	cmd := exec.Command("claude", cmdArgs...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// 设置环境变量
	env := os.Environ()
	// 这里可以添加配置中的环境变量
	cmd.Env = env

	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}