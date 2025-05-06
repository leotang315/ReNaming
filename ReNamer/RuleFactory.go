package ReNamer

import (
	"fmt"
)

// RuleFactory 规则工厂，用于创建各种重命名规则
type RuleFactory struct{}

// NewRuleFactory 创建新的规则工厂
func NewRuleFactory() *RuleFactory {
	return &RuleFactory{}
}

// AddPrefix 添加前缀
func (rf *RuleFactory) AddPrefix(prefix string) Rule {
	return Rule{
		Name:    "AddPrefix",
		Pattern: `^`,
		Replace: prefix,
	}
}

// AddSuffix 添加后缀
func (rf *RuleFactory) AddSuffix(suffix string) Rule {
	return Rule{
		Name:    "AddSuffix",
		Pattern: `$`,
		Replace: suffix,
	}
}

// AddAfterPattern 在匹配模式后添加内容
func (rf *RuleFactory) AddAfterPattern(pattern, content string) Rule {
	return Rule{
		Name:    "AddAfterPattern",
		Pattern: `(` + pattern + `)`,
		Replace: "${1}" + content,
	}
}

// AddBeforePattern 在匹配模式前添加内容
func (rf *RuleFactory) AddBeforePattern(pattern, content string) Rule {
	return Rule{
		Name:    "AddBeforePattern",
		Pattern: `(` + pattern + `)`,
		Replace: content + "${1}",
	}
}

// AddAtPosition 在指定位置添加内容
func (rf *RuleFactory) AddAtPosition(pos int, content string) Rule {
	return Rule{
		Name:    "AddAtPosition",
		Pattern: fmt.Sprintf(`^(.{%d})(.*)$`, pos),
		Replace: "${1}" + content + "${2}",
	}
}

// AddBeforeLastN 在末尾N个字符前添加内容
func (rf *RuleFactory) AddBeforeLastN(n int, content string) Rule {
	return Rule{
		Name:    "AddBeforeLastN",
		Pattern: fmt.Sprintf(`^(.*?)(.{%d})$`, n),
		Replace: "${1}" + content + "${2}",
	}
}

// RemovePattern 删除匹配的模式
func (rf *RuleFactory) RemovePattern(pattern string) Rule {
	return Rule{
		Name:    "RemovePattern",
		Pattern: pattern,
		Replace: "",
	}
}

// RemoveNumbers 删除数字
func (rf *RuleFactory) RemoveNumbers() Rule {
	return Rule{
		Name:    "RemoveNumbers",
		Pattern: `\d+`,
		Replace: "",
	}
}

// RemoveSpaces 删除空格
func (rf *RuleFactory) RemoveSpaces() Rule {
	return Rule{
		Name:    "RemoveSpaces",
		Pattern: `\s+`,
		Replace: "",
	}
}

// RemoveLetters 删除字母
func (rf *RuleFactory) RemoveLetters() Rule {
	return Rule{
		Name:    "RemoveLetters",
		Pattern: `[a-zA-Z]+`,
		Replace: "",
	}
}

// RemoveAtPosition 删除指定位置的字符
func (rf *RuleFactory) RemoveAtPosition(pos int) Rule {
	return Rule{
		Name:    "RemoveAtPosition",
		Pattern: fmt.Sprintf(`^(.{%d}).(.*)$`, pos),
		Replace: "${1}${2}",
	}
}

// RemoveFromEnd 从末尾删除N个字符
func (rf *RuleFactory) RemoveFromEnd(n int) Rule {
	return Rule{
		Name:    "RemoveFromEnd",
		Pattern: fmt.Sprintf(`^(.*?)(.{%d})$`, n),
		Replace: "${1}",
	}
}

// RemoveRange 删除指定范围的字符
func (rf *RuleFactory) RemoveRange(start, end int) Rule {
	return Rule{
		Name:    "RemoveRange",
		Pattern: fmt.Sprintf(`^(.{%d}).{%d}(.*)$`, start, end-start),
		Replace: "${1}${2}",
	}
}

// RemoveBetweenDelimiters 删除两个分隔符之间的内容，保留分隔符
func (rf *RuleFactory) RemoveBetweenDelimiters(startDelim, endDelim string) Rule {
	return Rule{
		Name:    "RemoveBetweenDelimiters",
		Pattern: startDelim + `[^` + endDelim + `]*` + endDelim,
		Replace: startDelim + endDelim,
	}
}

// RemoveWithDelimiters 删除两个分隔符之间的内容，包括分隔符
func (rf *RuleFactory) RemoveWithDelimiters(startDelim, endDelim string) Rule {
	return Rule{
		Name:    "RemoveWithDelimiters",
		Pattern: startDelim + `[^` + endDelim + `]*` + endDelim,
		Replace: "",
	}
}

// ReplacePattern 替换匹配的模式
func (rf *RuleFactory) ReplacePattern(oldPattern, newPattern string) Rule {
	return Rule{
		Name:    "ReplacePattern",
		Pattern: oldPattern,
		Replace: newPattern,
	}
}

// ReplaceSpaces 替换空格
func (rf *RuleFactory) ReplaceSpaces(replacement string) Rule {
	return Rule{
		Name:    "ReplaceSpaces",
		Pattern: `\s+`,
		Replace: replacement,
	}
}

// ReplaceNumbers 替换数字
func (rf *RuleFactory) ReplaceNumbers(replacement string) Rule {
	return Rule{
		Name:    "ReplaceNumbers",
		Pattern: `\d`,
		Replace: replacement,
	}
}

// ReplaceLetters 替换字母
func (rf *RuleFactory) ReplaceLetters(replacement string) Rule {
	return Rule{
		Name:    "ReplaceLetters",
		Pattern: `[a-zA-Z]`,
		Replace: replacement,
	}
}

// ReplaceAtPosition 替换指定位置的字符
func (rf *RuleFactory) ReplaceAtPosition(pos int, replacement string) Rule {
	return Rule{
		Name:    "ReplaceAtPosition",
		Pattern: fmt.Sprintf(`^(.{%d}).(.*)$`, pos),
		Replace: "${1}" + replacement + "${2}",
	}
}

// ReplaceRange 替换指定范围的字符
func (rf *RuleFactory) ReplaceRange(start, end int, replacement string) Rule {
	return Rule{
		Name:    "ReplaceRange",
		Pattern: fmt.Sprintf(`^(.{%d}).{%d}(.*)$`, start, end-start),
		Replace: "${1}" + replacement + "${2}",
	}
}

// ReplaceBetweenDelimiters 替换两个分隔符之间的内容
func (rf *RuleFactory) ReplaceBetweenDelimiters(startDelim, endDelim, replacement string) Rule {
	return Rule{
		Name:    "ReplaceBetweenDelimiters",
		Pattern: startDelim + `[^` + endDelim + `]*` + endDelim,
		Replace: startDelim + replacement + endDelim,
	}
}
