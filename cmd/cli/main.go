package main

import (
	"ReNaming/internal/renamer"
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

	// --- 1. Define command line parameters ---
	path := flag.String("path", "", "Comma-separated list of files and/or directories to process.")
	pattern := flag.String("pattern", "*", "Glob pattern to filter files in directories.")
	recursive := flag.Bool("recursive", false, "Process subdirectories recursively.")
	ruleFile := flag.String("ruleFile", "", "Path to the JSON config file containing renaming rules.")
	ruleJSON := flag.String("rule", "", "Renaming rules as a JSON array string. For single rule, wrap it in square brackets.")
	dryRun := flag.Bool("dry-run", false, "Preview changes without actually renaming files.")
	mappingFile := flag.String("mapping", "", "Path to the JSON file containing renaming mappings.")
	outputFile := flag.String("output", "", "Path to save the results as JSON file.")

	flag.Parse()

	// --- 2. Create Renamer ---
	reNamer := renamer.NewReNamer()
	reNamer.SetDryRun(*dryRun)

	// If mapping file is provided, use it directly for renaming
	if *mappingFile != "" {
		// Read mapping file
		data, err := os.ReadFile(*mappingFile)
		if err != nil {
			log.Fatalf("Error reading mapping file %s: %v", *mappingFile, err)
		}

		// Parse mapping data
		var mappings []renamer.ReNameResult
		if err := json.Unmarshal(data, &mappings); err != nil {
			log.Fatalf("Error parsing mapping file %s: %v", *mappingFile, err)
		}

		// Apply mappings
		results := reNamer.ApplyMapping(mappings, renamer.ModeError)

		// Print results
		resultsJSON, err := json.MarshalIndent(results, "", "    ")
		if err != nil {
			log.Printf("Error marshaling results: %v\n", err)
		} else {
			fmt.Printf("\n--- Results JSON ---\n%s\n", resultsJSON)
		}
		return
	}

	// If no mapping file is provided, continue with the original processing flow
	if *path == "" {
		log.Fatal("No path specified. Use -path to specify files and/or directories, or use -mapping to provide a mapping file.")
	}

	if *ruleFile != "" {
		// Load config from file
		data, err := os.ReadFile(*ruleFile)
		if err != nil {
			log.Fatalf("Error reading config file %s: %v", *ruleFile, err)
		}
		err = reNamer.LoadRule(data)
		if err != nil {
			log.Fatalf("Error parsing config file %s: %v", *ruleFile, err)
		}
		fmt.Printf("Loaded %d rules from config file: %s\n", len(reNamer.Rules), *ruleFile)
	} else if *ruleJSON != "" {
		// Load rules from command line (single or multiple)
		var rules []renamer.Rule
		if err := json.Unmarshal([]byte(*ruleJSON), &rules); err != nil {
			log.Fatalf("Error parsing -rule JSON: %v", err)
		}
		for _, rule := range rules {
			reNamer.AddRule(rule)
		}
		fmt.Printf("Loaded %d rules from command line\n", len(rules))
	}

	if len(reNamer.Rules) == 0 {
		log.Fatal("No renaming rules provided. Use -config or -rule.")
	}

	// --- 3. Get file list ---
	var filesToProcess []string

	// Process each path (can be file or directory)
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

		// Modify directory processing part
		if fileInfo.IsDir() {
			// Process directory
			fmt.Printf("Processing directory: %s (pattern: %s)\n", p, *pattern)
			// Use filepath.WalkDir instead of os.ReadDir
			err = filepath.WalkDir(p, func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					log.Printf("Warning: Error accessing %s: %v\n", path, err)
					return nil // Continue processing other files
				}

				// If directory and not recursive mode, skip subdirectory
				if d.IsDir() {
					if !*recursive && path != p {
						return filepath.SkipDir
					}
					return nil
				}

				// Process file
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
			// Process single file
			filesToProcess = append(filesToProcess, p)
			fmt.Printf("Processing file: %s\n", p)
		}
	}

	if len(filesToProcess) == 0 {
		fmt.Println("No files found to process.")
		return
	}

	// --- 4. Apply rename rules ---
	reNamer.AddFiles(filesToProcess)
	reNamer.SetDryRun(*dryRun)
	results := reNamer.ApplyBatch()

	// Print results in JSON format
	resultsJSON, err := json.MarshalIndent(results, "", "    ")
	if err != nil {
		log.Printf("Error marshaling results: %v\n", err)
	} else {
		fmt.Printf("\n--- Results JSON ---\n%s\n", resultsJSON)

		// 如果指定了输出文件，则将结果保存到文件
		if *outputFile != "" {
			err := os.WriteFile(*outputFile, resultsJSON, 0644)
			if err != nil {
				log.Printf("Error writing results to file %s: %v\n", *outputFile, err)
			} else {
				fmt.Printf("Results have been saved to: %s\n", *outputFile)
			}
		}
	}
}
