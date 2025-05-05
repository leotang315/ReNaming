package main

import (
	"ReNaming/ReNamer"
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {

	// --- 1. 定义命令行参数 ---
	path := flag.String("path", "", "Comma-separated list of files and/or directories to process.")
	pattern := flag.String("pattern", "*", "Glob pattern to filter files in directories.")
	recursive := flag.Bool("recursive", false, "Process subdirectories recursively.")
	ruleFile := flag.String("ruleFile", "", "Path to the JSON config file containing renaming rules.")
	ruleJSON := flag.String("rule", "", "Renaming rules as a JSON array string. For single rule, wrap it in square brackets.")
	dryRun := flag.Bool("dry-run", false, "Preview changes without actually renaming files.")

	flag.Parse()

	if *path == "" {
		log.Fatal("No path specified. Use -path to specify files and/or directories.")
	}

	// --- 2. 创建和配置 Renamer ---
	renamer := ReNamer.NewRenamer()

	if *ruleFile != "" {
		// 从文件加载配置
		data, err := os.ReadFile(*ruleFile)
		if err != nil {
			log.Fatalf("Error reading config file %s: %v", *ruleFile, err)
		}
		err = renamer.LoadRule(data)
		if err != nil {
			log.Fatalf("Error parsing config file %s: %v", *ruleFile, err)
		}
		fmt.Printf("Loaded %d rules from config file: %s\n", len(renamer.Rules), *ruleFile)
	} else if *ruleJSON != "" {
		// 从命令行加载规则（单个或多个）
		var rules []ReNamer.Rule
		if err := json.Unmarshal([]byte(*ruleJSON), &rules); err != nil {
			log.Fatalf("Error parsing -rule JSON: %v", err)
		}
		for _, rule := range rules {
			renamer.AddRule(rule)
		}
		fmt.Printf("Loaded %d rules from command line\n", len(rules))
	}

	if len(renamer.Rules) == 0 {
		log.Fatal("No renaming rules provided. Use -config or -rule.")
	}

	// --- 3. 获取文件列表 ---
	var filesToProcess []string

	// 处理每个路径（可以是文件或目录）
	paths := strings.Split(*path, ",")
	for _, p := range paths {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}

		fileInfo, err := os.Stat(p)
		if err != nil {
			log.Printf("Warning: Cannot access %s: %v\n", p, err)
			continue
		}

		// 修改目录处理部分
		if fileInfo.IsDir() {
			// 处理目录
			fmt.Printf("Processing directory: %s (pattern: %s)\n", p, *pattern)
			// 使用 filepath.WalkDir 替代 os.ReadDir
			err = filepath.WalkDir(p, func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					log.Printf("Warning: Error accessing %s: %v\n", path, err)
					return nil // 继续处理其他文件
				}

				// 如果是目录且不是递归模式，跳过子目录
				if d.IsDir() {
					if !*recursive && path != p {
						return filepath.SkipDir
					}
					return nil
				}

				// 处理文件
				match, _ := filepath.Match(*pattern, d.Name())
				if match {
					filesToProcess = append(filesToProcess, path)
				}
				return nil
			})

			if err != nil {
				log.Printf("Warning: Error walking directory %s: %v\n", p, err)
			}
		} else {
			// 处理单个文件
			filesToProcess = append(filesToProcess, p)
			fmt.Printf("Processing file: %s\n", p)
		}
	}

	if len(filesToProcess) == 0 {
		fmt.Println("No files found to process.")
		return
	}

	// --- 4. 应用重命名规则 ---
	renamer.AddFiles(filesToProcess)
	renamer.SetDryRun(*dryRun)
	results := renamer.ApplyBatch()

	// 打印 results 的 JSON 格式
	resultsJSON, err := json.MarshalIndent(results, "", "    ")
	if err != nil {
		log.Printf("Error marshaling results: %v\n", err)
	} else {
		fmt.Printf("\n--- Results JSON ---\n%s\n", resultsJSON)
	}

}
