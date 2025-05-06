package ReNamer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

// ReNameResult represents the result of a rename operation with error handling
type ReNameResult struct {
	OldPath string       `json:"oldPath"`
	NewPath string       `json:"newPath"`
	Status  ReNameStatus `json:"status"`            // Status: success, error, pending
	Message string       `json:"message,omitempty"` // Detailed message, only shown when needed
}

// ReNamer represents the file renaming manager
type ReNamer struct {
	Rules            []Rule         `json:"operations"`
	FileList         []string       `json:"files"`
	DryRun           bool           `json:"dryRun"`           // Whether in preview mode
	Mappings         []ReNameResult `json:"mappings"`         // Results of nename mappings
	ProcessExtension bool           `json:"processExtension"` // Whether to process file extension
}

func NewReNamer() *ReNamer {
	return &ReNamer{
		Rules:            make([]Rule, 0),
		FileList:         make([]string, 0),
		DryRun:           false,
		Mappings:         make([]ReNameResult, 0),
		ProcessExtension: false, // 默认不处理扩展名
	}
}

// SetProcessExtension 设置是否处理文件扩展名
func (r *ReNamer) SetProcessExtension(process bool) {
	r.ProcessExtension = process
}

// generateSingleMapping generates rename mapping for a single file
func (r *ReNamer) generateSingleMapping(path string) ReNameResult {
	result := ReNameResult{
		OldPath: path,
		Status:  StatusPending,
	}

	if path == "" {
		result.Status = StatusError
		result.Message = "Empty file path"
		return result
	}

	// Split path and filename
	dir, srcName := filepath.Split(path)
	if srcName == "" {
		result.Status = StatusError
		result.Message = "Invalid file path"
		return result
	}

	ext := filepath.Ext(srcName)
	src := srcName
	dst := srcName

	if !r.ProcessExtension {
		src = srcName[:len(srcName)-len(ext)]
	}

	for _, op := range r.Rules {
		var err error
		dst, err = op.Apply(src)
		if err != nil {
			result.Status = StatusError
			result.Message = err.Error()
			return result
		}
		if dst == "" {
			result.Status = StatusError
			result.Message = "Generated filename is invalid"
			return result
		}
	}
	if !r.ProcessExtension {
		dst = dst + ext
	}
	result.NewPath = filepath.Join(dir, dst)

	return result
}

func (r *ReNamer) SetDryRun(dryRun bool) {
	r.DryRun = dryRun
}

func (r *ReNamer) AddFiles(files []string) {
	r.FileList = append(r.FileList, files...)
}

func (r *ReNamer) RemoveFile(file string) bool {
	for i, f := range r.FileList {
		if f == file {
			r.FileList = append(r.FileList[:i], r.FileList[i+1:]...)
			return true
		}
	}
	return false
}

func (r *ReNamer) ClearFiles() {
	r.FileList = make([]string, 0)
}

func (r *ReNamer) GetFiles() []string {
	files := make([]string, len(r.FileList))
	copy(files, r.FileList)
	return files
}

func (r *ReNamer) AddRule(rule Rule) string {
	if rule.ID == "" {
		rule.ID = uuid.New().String()
	}
	r.Rules = append(r.Rules, rule)
	return rule.ID
}

func (r *ReNamer) RemoveRuleByID(id string) bool {
	for i, op := range r.Rules {
		if op.ID == id {
			r.Rules = append(r.Rules[:i], r.Rules[i+1:]...)
			return true
		}
	}
	return false
}

func (r *ReNamer) RemoveRuleByName(name string) int {
	newOperations := make([]Rule, 0)
	removedCount := 0

	for _, op := range r.Rules {
		if op.Name != name {
			newOperations = append(newOperations, op)
		} else {
			removedCount++
		}
	}

	r.Rules = newOperations
	return removedCount
}

func (r *ReNamer) GetRuleByID(id string) *Rule {
	for _, op := range r.Rules {
		if op.ID == id {
			return &op
		}
	}
	return nil
}

func (r *ReNamer) SaveRule() ([]byte, error) {
	return json.Marshal(r.Rules)
}

func (r *ReNamer) LoadRule(data []byte) error {
	return json.Unmarshal(data, &r.Rules)
}

// ApplyBatch processes all files in the internal file list in batch
func (r *ReNamer) ApplyBatch() []ReNameResult {
	// 生成映射
	r.Mappings = make([]ReNameResult, len(r.FileList))
	for i, filename := range r.FileList {
		r.Mappings[i] = r.generateSingleMapping(filename)
	}

	results := r.ApplyMapping(r.Mappings, ModeNormal)

	return results
}

// ApplyMapping executes the rename mapping list based on the specified mode
// Parameters:
//   - mappings: List of rename operations to be executed
//   - mode: Operation mode that determines how mappings are processed:
//     ModeNormal: Skip mappings with error status
//     ModeError: Only retry mappings with error status
//     ModeUndo: Reverse the rename operation by swapping OldPath and NewPath
//
// Returns: List of ReNameResult containing the execution results
func (r *ReNamer) ApplyMapping(mappings []ReNameResult, mode ReNameMode) []ReNameResult {
	// 如果是预览模式，直接返回映射结果
	if r.DryRun {
		return mappings
	}

	// 执行实际重命名操作
	results := make([]ReNameResult, len(mappings))
	copy(results, mappings)

	for i, mapping := range results {
		switch mode {
		case ModeNormal:
			if mapping.Status == StatusError {
				continue
			}
		case ModeError:
			if mapping.Status != StatusError {
				continue
			}
		case ModeUndo:
			// 执行回退操作，交换新旧路径
			oldPath := mapping.OldPath
			mapping.OldPath = mapping.NewPath
			mapping.NewPath = oldPath
		default:
			results[i].Status = StatusError
			results[i].Message = "无效的操作模式"
			continue
		}

		// 检查路径有效性
		if mapping.OldPath == "" || mapping.NewPath == "" {
			results[i].Status = StatusError
			results[i].Message = "无效的文件路径"
			continue
		}

		// 如果新旧路径相同，标记为成功并跳过
		if mapping.OldPath == mapping.NewPath {
			results[i].Status = StatusSuccess // 使用枚举值
			continue
		}

		err := os.Rename(mapping.OldPath, mapping.NewPath)
		if err != nil {
			results[i].Status = StatusError
			results[i].Message = fmt.Sprintf("Rename failed: %v", err)
		} else {
			results[i].Status = StatusSuccess
		}
	}

	return results
}
