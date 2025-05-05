package ReNamer

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// Rule 重命名规则
type Rule struct {
	ID      string `json:"id"`      // 唯一标识符
	Name    string `json:"name"`    // 操作名称
	Pattern string `json:"pattern"` // 匹配模式
	Replace string `json:"Replace"` // 替换模板
}

// Apply 应用规则到文件名，返回新文件名和错误
func (r Rule) Apply(filename string) (string, error) {
	template := r.processPlaceholders(r.Replace)

	// 安全地编译正则表达式
	re, err := regexp.Compile(r.Pattern)
	if err != nil {
		return filename, fmt.Errorf("无效的正则表达式 '%s': %v", r.Pattern, err)
	}

	return re.ReplaceAllString(filename, template), nil
}

// processPlaceholders 处理替换模板中的占位符
func (r Rule) processPlaceholders(text string) string {
	// 处理日期时间
	text = strings.ReplaceAll(text, "{date}", time.Now().Format("2006-01-02"))
	text = strings.ReplaceAll(text, "{time}", time.Now().Format("15:04:05"))
	text = strings.ReplaceAll(text, "{datetime}", time.Now().Format("2006-01-02_15:04:05"))

	// 处理其他占位符...

	return text
}
