package Renamer

import (
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

// Apply 应用规则到文件名
func (r Rule) Apply(filename string) string {
	template := r.processPlaceholders(r.Replace)
	re := regexp.MustCompile(r.Pattern)
	return re.ReplaceAllString(filename, template)
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