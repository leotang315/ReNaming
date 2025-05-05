package ReNamer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

// ReNameResult 带错误处理结果
type ReNameResult struct {
	OldPath string `json:"oldPath"`
	NewPath string `json:"newPath"`
	Status  string `json:"status"`            // 状态：success, error, undo
	Message string `json:"message,omitempty"` // 详细信息，仅在需要时显示
}

// ReNamer 重命名器
type ReNamer struct {
	Rules    []Rule         `json:"operations"`
	FileList []string       `json:"files"`
	DryRun   bool           `json:"dryRun"`   // 是否为预览模式
	Mappings []ReNameResult `json:"mappings"` // 重命名映射结果
}

func NewRenamer() *ReNamer {
	return &ReNamer{
		Rules:    make([]Rule, 0),
		FileList: make([]string, 0),
		DryRun:   false,
		Mappings: make([]ReNameResult, 0),
	}
}

// generateSingleMapping 生成单个文件的重命名映射
func (r *ReNamer) generateSingleMapping(path string) ReNameResult {
	result := ReNameResult{
		OldPath: path,
		Status:  "undo",
	}

	if path == "" {
		result.Status = "error"
		result.Message = "空文件路径"
		return result
	}

	// 分离路径和文件名
	dir, srcName := filepath.Split(path)
	if srcName == "" {
		result.Status = "error"
		result.Message = "无效的文件路径"
		return result
	}

	// 只处理文件名部分
	dstName := srcName
	for _, op := range r.Rules {
		var err error
		dstName, err = op.Apply(dstName)
		if err != nil {
			result.Status = "error"
			result.Message = err.Error()
			return result
		}
		if dstName == "" {
			result.Status = "error"
			result.Message = "生成的文件名无效"
			return result
		}
	}

	// 组合新的完整路径
	result.NewPath = filepath.Join(dir, dstName)
	return result
}

// SetDryRun 设置是否为预览模式
func (r *ReNamer) SetDryRun(dryRun bool) {
	r.DryRun = dryRun
}

// ApplyBatch 批量处理内部文件列表中的所有文件
func (r *ReNamer) ApplyBatch() []ReNameResult {
	// 生成映射
	r.Mappings = make([]ReNameResult, len(r.FileList))
	for i, filename := range r.FileList {
		r.Mappings[i] = r.generateSingleMapping(filename)
	}

	// 如果是预览模式，直接返回映射结果
	if r.DryRun {
		return r.Mappings
	}

	// 执行实际重命名操作
	results := make([]ReNameResult, len(r.Mappings))
	copy(results, r.Mappings)

	for i, mapping := range results {
		if mapping.Status == "error" {
			continue
		}
		if mapping.OldPath != mapping.NewPath {
			err := os.Rename(mapping.OldPath, mapping.NewPath)
			if err != nil {
				results[i].Status = "error"
				results[i].Message = fmt.Sprintf("重命名失败：%v", err)
			} else {
				results[i].Status = "success"
			}
		}
	}

	return results
}

// AddFiles 添加文件到处理列表
func (r *ReNamer) AddFiles(files []string) {
	r.FileList = append(r.FileList, files...)
}

// RemoveFile 从处理列表中移除指定文件
func (r *ReNamer) RemoveFile(file string) bool {
	for i, f := range r.FileList {
		if f == file {
			r.FileList = append(r.FileList[:i], r.FileList[i+1:]...)
			return true
		}
	}
	return false
}

// ClearFiles 清空文件列表
func (r *ReNamer) ClearFiles() {
	r.FileList = make([]string, 0)
}

// GetFiles 获取当前的文件列表
func (r *ReNamer) GetFiles() []string {
	files := make([]string, len(r.FileList))
	copy(files, r.FileList)
	return files
}

// Add 添加规则并返回规则ID
func (r *ReNamer) AddRule(rule Rule) string {
	if rule.ID == "" {
		rule.ID = uuid.New().String()
	}
	r.Rules = append(r.Rules, rule)
	return rule.ID
}

// RemoveRuleByID 通过ID删除操作
func (r *ReNamer) RemoveRuleByID(id string) bool {
	for i, op := range r.Rules {
		if op.ID == id {
			r.Rules = append(r.Rules[:i], r.Rules[i+1:]...)
			return true
		}
	}
	return false
}

// RemoveRuleByName 通过名称删除所有匹配的操作
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

// GetRuleByID 通过ID获取操作
func (r *ReNamer) GetRuleByID(id string) *Rule {
	for _, op := range r.Rules {
		if op.ID == id {
			return &op
		}
	}
	return nil
}

// SaveRule 保存规则配置到JSON
func (r *ReNamer) SaveRule() ([]byte, error) {
	return json.Marshal(r.Rules)
}

// LoadRule 从JSON加载规则配置
func (r *ReNamer) LoadRule(data []byte) error {
	return json.Unmarshal(data, &r.Rules)
}
