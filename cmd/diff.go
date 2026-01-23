package cmd

import (
	"fmt"
	"strings"

	"github.com/fiftyk/claude-switcher/internal/profile"
)

// FieldDiff 表示单个字段的差异
type FieldDiff struct {
	Field  string
	Value1 string
	Value2 string
}

// ProfileDiff 表示两个配置的比较结果
type ProfileDiff struct {
	HasDifferences bool
	Profile1       string
	Profile2       string
	Differences    []FieldDiff
}

// DiffProfiles 比较两个配置
func DiffProfiles(p1, p2 *profile.Profile) *ProfileDiff {
	diff := &ProfileDiff{
		HasDifferences: false,
		Profile1:       p1.Name,
		Profile2:       p2.Name,
		Differences:    []FieldDiff{},
	}

	// 比较各字段
	if p1.AuthToken != p2.AuthToken {
		diff.Differences = append(diff.Differences, FieldDiff{
			Field:  "AuthToken",
			Value1: maskToken(p1.AuthToken),
			Value2: maskToken(p2.AuthToken),
		})
	}

	if p1.BaseURL != p2.BaseURL {
		diff.Differences = append(diff.Differences, FieldDiff{
			Field:  "BaseURL",
			Value1: p1.BaseURL,
			Value2: p2.BaseURL,
		})
	}

	if p1.HTTPProxy != p2.HTTPProxy {
		diff.Differences = append(diff.Differences, FieldDiff{
			Field:  "HTTPProxy",
			Value1: p1.HTTPProxy,
			Value2: p2.HTTPProxy,
		})
	}

	if p1.HTTPSProxy != p2.HTTPSProxy {
		diff.Differences = append(diff.Differences, FieldDiff{
			Field:  "HTTPSProxy",
			Value1: p1.HTTPSProxy,
			Value2: p2.HTTPSProxy,
		})
	}

	if p1.Model != p2.Model {
		diff.Differences = append(diff.Differences, FieldDiff{
			Field:  "Model",
			Value1: p1.Model,
			Value2: p2.Model,
		})
	}

	// 比较自定义环境变量
	for k, v1 := range p1.EnvVars {
		if v2, ok := p2.EnvVars[k]; !ok || v1 != v2 {
			diff.Differences = append(diff.Differences, FieldDiff{
				Field:  "EnvVar:" + k,
				Value1: v1,
				Value2: v2,
			})
		}
	}

	// 检查 p2 中额外的环境变量
	for k, v2 := range p2.EnvVars {
		if _, ok := p1.EnvVars[k]; !ok {
			diff.Differences = append(diff.Differences, FieldDiff{
				Field:  "EnvVar:" + k,
				Value1: "",
				Value2: v2,
			})
		}
	}

	diff.HasDifferences = len(diff.Differences) > 0
	return diff
}

// FormatDiffOutput 格式化差异输出
func FormatDiffOutput(diff *ProfileDiff) string {
	var sb strings.Builder

	sb.WriteString("\n")
	sb.WriteString(strings.Repeat("=", 60) + "\n")
	sb.WriteString(fmt.Sprintf("配置比较: %s vs %s\n", diff.Profile1, diff.Profile2))
	sb.WriteString(strings.Repeat("=", 60) + "\n\n")

	if !diff.HasDifferences {
		sb.WriteString("  ✓ 两个配置完全相同\n\n")
		return sb.String()
	}

	sb.WriteString(fmt.Sprintf("  发现 %d 处差异:\n\n", len(diff.Differences)))

	for i, d := range diff.Differences {
		sb.WriteString(fmt.Sprintf("  %d. %s\n", i+1, d.Field))
		if d.Value1 == "" {
			sb.WriteString(fmt.Sprintf("     - 新增: %s\n", d.Value2))
		} else if d.Value2 == "" {
			sb.WriteString(fmt.Sprintf("     - 移除: %s\n", d.Value1))
		} else {
			sb.WriteString(fmt.Sprintf("     - 之前: %s\n", d.Value1))
			sb.WriteString(fmt.Sprintf("     + 之后: %s\n", d.Value2))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// PrintDiff 打印配置差异
func PrintDiff(profilesDir, name1, name2 string) error {
	p1, err := profile.LoadProfile(profilesDir, name1)
	if err != nil {
		return fmt.Errorf("无法加载配置 '%s': %w", name1, err)
	}

	p2, err := profile.LoadProfile(profilesDir, name2)
	if err != nil {
		return fmt.Errorf("无法加载配置 '%s': %w", name2, err)
	}

	diff := DiffProfiles(p1, p2)
	fmt.Print(FormatDiffOutput(diff))

	return nil
}

// PrintDiffHelp 打印比较帮助信息
func PrintDiffHelp() {
	fmt.Println("\n配置比较用法:")
	fmt.Println("  claude-switcher --diff <配置1> <配置2>")
	fmt.Println()
	fmt.Println("示例:")
	fmt.Println("  claude-switcher --diff work personal")
	fmt.Println()
}
