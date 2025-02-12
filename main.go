package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Config 全局配置结构体，包含所有命令行参数和运行时配置
type Config struct {
	InputFolder  string        // 输入文件夹路径
	Pattern      string        // 文件匹配模式（例如 *.txt）
	Recursive    bool          // 是否递归处理子目录
	NameTemplate string        // 新文件名模板
	DryRun       bool          // 试运行模式（只显示不执行）
	GenerateMap  string        // 生成映射文件路径（用于保存重命名计划）
	ApplyMap     string        // 应用映射文件路径（用于执行预存的重命名计划）
	DateFormat   string        // 日期格式（用于 {date} 占位符）
	TimeFormat   string        // 时间格式（用于 {time} 占位符）
	Index        int           // 全局文件索引计数器
	Mappings     []FileMapping // 文件路径映射关系集合
}

// FileMapping 文件路径映射记录，记录原始路径和新路径的对应关系
type FileMapping struct {
	OldPath string
	NewPath string
}

// 定义占位符处理函数类型
type placeholderHandler func(placeholder string, data map[string]string) string

// 占位符处理器映射
var placeholderHandlers = map[string]placeholderHandler{
	"name":     handleName,
	"ext":      handleExt,
	"lower":    handleLower,
	"upper":    handleUpper,
	"title":    handleTitle,
	"date":     handleDate,
	"time":     handleTime,
	"datetime": handleDatetime,
}
var config Config

// 主函数
func main() {

	flag.StringVar(&config.InputFolder, "f", "", "Input folder path (required)")
	flag.StringVar(&config.NameTemplate, "n", "", "New name template (required)")
	flag.StringVar(&config.Pattern, "p", "*", "File pattern (e.g., *.txt)")
	flag.BoolVar(&config.Recursive, "r", false, "Recursively process subfolders")
	flag.BoolVar(&config.DryRun, "d", false, "Dry run (show changes without renaming)")
	flag.StringVar(&config.GenerateMap, "g", "", "Generate mapping file (dry run)")
	flag.StringVar(&config.ApplyMap, "a", "", "Apply mapping from file")
	flag.StringVar(&config.DateFormat, "df", "YYYY-MM-DD", "Date format for {date} placeholder")
	flag.StringVar(&config.TimeFormat, "tf", "HH:MM:SS", "Time format for {time} placeholder")

	flag.Parse()

	// 检查必要参数
	if config.InputFolder == "" || config.NameTemplate == "" {
		flag.Usage()
		return
	}

	// 检查互斥参数
	if config.GenerateMap != "" && config.ApplyMap != "" {
		logError("Cannot use -g and -a together")
		return
	}

	// 初始化
	config.Index = 0
	config.Mappings = make([]FileMapping, 0)

	// 构建映射
	if config.ApplyMap != "" {
		getMappingFromFile()
	} else {
		getMappingFromProcess()

	}

	// 应用映射
	rename()
}

// processFile 处理单个文件，生成新文件名并记录映射关系
// filePath: 当前处理的文件完整路径
func processFile(filePath string) {
	fileName := filepath.Base(filePath)
	fileExt := filepath.Ext(fileName)
	fileNameWithoutExt := strings.TrimSuffix(fileName, fileExt)

	// 占位符数据
	data := map[string]string{
		"filePath": filePath,
		"fileName": fileNameWithoutExt,
		"fileExt":  strings.TrimPrefix(fileExt, "."),
	}

	// 处理模板
	newName := processTemplate(config.NameTemplate, data)
	newPath := filepath.Join(config.InputFolder, newName)

	// 生成映射
	config.Mappings = append(config.Mappings, FileMapping{
		OldPath: filePath,
		NewPath: newPath,
	})

	// 更新序号
	config.Index++
}

// processTemplate 处理文件名模板，替换所有占位符
// template: 原始模板字符串
// data: 包含文件相关数据的字典
// 返回值: 处理后的文件名
func processTemplate(template string, data map[string]string) string {

	// 替换简单占位符
	for key, handler := range placeholderHandlers {
		placeholder := "{" + key + "}"
		if strings.Contains(template, placeholder) {
			template = strings.ReplaceAll(template, placeholder, handler(placeholder, data))
		}
	}

	// 处理复杂占位符（如 {regex:...}）
	template = processComplexPlaceholders(template, data)

	return template
}

// processComplexPlaceholders 处理复杂格式占位符
// 支持以下类型占位符：
// - {index:起始值:补零位数} 自动递增索引
// - {split:分隔符:索引} 字符串分割
// - {slice:起始位置:结束位置} 字符串截取
// - {replace:旧字符串:新字符串} 字符串替换
// - {regex:正则表达式:捕获组} 正则表达式提取
func processComplexPlaceholders(template string, data map[string]string) string {

	// 分段占位符 {index:...}
	indexRegex := regexp.MustCompile(`{index(?::(\d+)(?::(\d+))?)?}`)
	template = indexRegex.ReplaceAllStringFunc(template, func(match string) string {
		parts := indexRegex.FindStringSubmatch(match)
		placeholder := parts[0]
		indexStart := "1"
		indexZeroPadding := "0"
		if len(parts) > 1 && parts[1] != "" {
			indexStart = parts[1]
		}
		if len(parts) > 2 && parts[2] != "" {
			indexZeroPadding = parts[2]
		}
		data["indexStart"] = indexStart
		data["indexZeroPadding"] = indexZeroPadding
		return handleIndex(placeholder, data)
	})

	// 分段占位符 {split:...}
	splitRegex := regexp.MustCompile(`{split:([^:]+):(\d+)}`)
	template = splitRegex.ReplaceAllStringFunc(template, func(match string) string {
		parts := splitRegex.FindStringSubmatch(match)
		placeholder := parts[0]
		data["sep"] = parts[1]
		data["index"] = parts[2]
		return handleSplit(placeholder, data)
	})

	// 截取占位符 {slice:...}
	sliceRegex := regexp.MustCompile(`{slice:(\d+):(\d+)}`)
	template = sliceRegex.ReplaceAllStringFunc(template, func(match string) string {
		parts := sliceRegex.FindStringSubmatch(match)
		placeholder := parts[0]
		data["start"] = parts[1]
		data["end"] = parts[2]
		return handleSlice(placeholder, data)
	})

	// 替换占位符 {replace:...}
	replaceRegex := regexp.MustCompile(`{replace:([^:]+):([^}]+)}`)
	template = replaceRegex.ReplaceAllStringFunc(template, func(match string) string {
		parts := replaceRegex.FindStringSubmatch(match)
		placeholder := parts[0]
		data["old"] = parts[1]
		data["new"] = parts[2]
		return handleReplace(placeholder, data)
	})
	// 正则表达式占位符 {regex:...}
	regex := regexp.MustCompile(`{regex:([^:]+):(\d+)}`)
	template = regex.ReplaceAllStringFunc(template, func(match string) string {
		parts := regex.FindStringSubmatch(match)
		placeholder := parts[0]
		data["expr"] = parts[1]
		data["group"] = parts[2]
		return handleRegex(placeholder, data)
	})
	return template
}

// handleIndex 处理索引占位符 {index:start:padding}
// placeholder: 完整占位符字符串（未使用）
// data: 包含 indexStart 和 indexZeroPadding 的上下文数据
// 返回值: 格式化后的索引字符串
func handleIndex(placeholder string, data map[string]string) string {
	indexZeroPadding, _ := strconv.Atoi(data["indexZeroPadding"])
	indexStart, _ := strconv.Atoi(data["indexStart"])
	return fmt.Sprintf("%0*d", indexZeroPadding, config.Index+indexStart)
}

// 处理文件名占位符 {name}
func handleName(placeholder string, data map[string]string) string {
	return data["fileName"]
}

// 处理扩展名占位符 {ext}
func handleExt(placeholder string, data map[string]string) string {
	return data["fileExt"]
}

// 处理日期占位符 {date}
func handleDate(placeholder string, data map[string]string) string {
	format := convertDateFormat(config.DateFormat)
	return time.Now().Format(format)
}

// 处理时间占位符 {time}
func handleTime(placeholder string, data map[string]string) string {
	format := convertDateFormat(config.TimeFormat)

	return time.Now().Format(format)
}

// 处理时间占位符 {datetime}
func handleDatetime(placeholder string, data map[string]string) string {
	format := convertDateFormat(config.DateFormat + "_" + config.TimeFormat)
	return time.Now().Format(format)
}

// handleRegex 处理正则表达式占位符 {regex:pattern:group}
// placeholder: 完整占位符字符串（未使用）
// data: 包含 expr（正则表达式）和 group（捕获组号）
// 返回值: 匹配到的字符串或空字符串
func handleRegex(placeholder string, data map[string]string) string {
	expr := data["expr"]
	group, _ := strconv.Atoi(data["group"])
	re := regexp.MustCompile(expr)
	matches := re.FindStringSubmatch(data["fileName"])
	if len(matches) > group {
		return matches[group]
	}
	return ""
}

// 处理分段占位符 {split:SEP:INDEX}
func handleSplit(placeholder string, data map[string]string) string {
	sep := data["sep"]
	index, _ := strconv.Atoi(data["index"])
	parts := strings.Split(data["fileName"], sep)
	if index >= 0 && index < len(parts) {
		return parts[index]
	}
	return ""
}

// 处理截取占位符 {slice:START:END}
func handleSlice(placeholder string, data map[string]string) string {
	start, _ := strconv.Atoi(data["start"])
	end, _ := strconv.Atoi(data["end"])
	if start < 0 {
		start = 0
	}

	if end > len(data["fileName"]) {
		end = len(data["fileName"])
	}
	return data["fileName"][start:end]
}

// 处理替换占位符 {replace:OLD:NEW}
func handleReplace(placeholder string, data map[string]string) string {
	old := data["old"]
	new := data["new"]
	return strings.ReplaceAll(data["fileName"], old, new)
}

// 处理小写占位符 {lower}
func handleLower(placeholder string, data map[string]string) string {
	return strings.ToLower(data["fileName"])
}

// 处理大写占位符 {upper}
func handleUpper(placeholder string, data map[string]string) string {
	return strings.ToUpper(data["fileName"])
}

// 处理标题格式占位符 {title}
func handleTitle(placeholder string, data map[string]string) string {
	return strings.Title(data["fileName"])
}

// Logging functions
func logVerbose(format string, args ...interface{}) {
	fmt.Printf(format+"\n", args...)
}

func logError(format string, args ...interface{}) {
	fmt.Printf("Error: "+format+"\n", args...)
}

// convertDateFormat 将用户友好的日期格式转换为 Go 的格式
// userFormat: 用户输入的日期格式（支持 YYYY, MM, DD, HH, mm, ss）
// 返回值: Go 语言兼容的日期格式字符串
func convertDateFormat(userFormat string) string {
	// 替换常见的日期格式符号
	userFormat = strings.ReplaceAll(userFormat, "YYYY", "2006")
	userFormat = strings.ReplaceAll(userFormat, "MM", "01")
	userFormat = strings.ReplaceAll(userFormat, "DD", "02")
	userFormat = strings.ReplaceAll(userFormat, "HH", "15")
	userFormat = strings.ReplaceAll(userFormat, "mm", "04")
	userFormat = strings.ReplaceAll(userFormat, "ss", "05")
	return userFormat
}

// rename 执行重命名操作，根据不同模式：
// - 生成映射文件模式：将映射关系写入CSV
// - 试运行模式：在控制台显示变更计划
// - 正常模式：实际执行文件重命名
func rename() {

	if config.GenerateMap != "" {
		writeMappingsToFile()
		return
	}

	if config.DryRun {
		writeMappingsToConsole()
		return
	}

	writeMappingsToReal()
}

// 写入映射关系（处理完成后统一调用）
func writeMappingsToFile() {
	f, err := os.Create(config.GenerateMap)
	if err != nil {
		logError("Error creating mapping file: %v", err)
		return
	}
	defer f.Close()

	writer := csv.NewWriter(f)
	defer writer.Flush()

	// 写入CSV头
	_ = writer.Write([]string{"old_path", "new_path"})

	for _, m := range config.Mappings {
		if err := writer.Write([]string{m.OldPath, m.NewPath}); err != nil {
			logError("Error writing mapping: %v", err)
		}
	}
}

func writeMappingsToConsole() {
	for _, m := range config.Mappings {
		logVerbose("Dry run: would rename %s to %s", m.OldPath, m.NewPath)
	}
}

func writeMappingsToReal() {
	for _, m := range config.Mappings {
		// 处理文件名冲突
		if _, err := os.Stat(m.NewPath); err == nil {
			logVerbose("Skipping conflicting file: %s", m.NewPath)
			return
		}
		logVerbose("Renaming %s to %s", m.OldPath, m.NewPath)
		if err := os.Rename(m.OldPath, m.NewPath); err != nil {
			logError("Error renaming file: %v", err)
		}
	}
}

// getMappingFromFile 从CSV文件加载映射关系
// 文件格式：old_path,new_path
func getMappingFromFile() {
	f, err := os.Open(config.ApplyMap)
	if err != nil {
		logError("Error reading mapping file: %v", err)
		return
	}
	defer f.Close()

	reader := csv.NewReader(f)

	// 跳过标题行
	if _, err := reader.Read(); err != nil {
		logError("Error reading header: %v", err)
		return
	}

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			logError("Error parsing mapping file: %v", err)
			continue
		}
		if len(record) != 2 {
			continue
		}

		x := FileMapping{
			OldPath: record[0],
			NewPath: record[1],
		}
		config.Mappings = append(config.Mappings, x)
	}
}

func getMappingFromProcess() {
	// 遍历输入文件夹
	err := filepath.Walk(config.InputFolder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			// 检查文件模式匹配
			if matched, _ := filepath.Match(config.Pattern, filepath.Base(path)); matched {
				processFile(path)
			}
		} else if !config.Recursive && path != config.InputFolder {
			// 如果不递归，跳过子文件夹
			return filepath.SkipDir
		}
		return nil
	})

	logError("Error Walk : %v", err)
}
