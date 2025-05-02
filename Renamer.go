package Renamer

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

// RenameResult 带错误处理结果
type RenameResult struct {
	OldName string `json:"old_name"`
	NewName string `json:"new_name"`
	Error   string `json:"error,omitempty"` // 改为string类型，并添加omitempty标签
}

// Renamer 重命名器
type Renamer struct {
	Rules []Rule `json:"operations"`
}

func NewRenamer() *Renamer {
	return &Renamer{
		Rules: make([]Rule, 0),
	}
}

// Add 添加规则并返回规则ID
func (r *Renamer) Add(rule Rule) string {
	if rule.ID == "" {
		rule.ID = uuid.New().String()
	}
	r.Rules = append(r.Rules, rule)
	return rule.ID
}

// RemoveByID 通过ID删除操作
func (r *Renamer) RemoveByID(id string) bool {
	for i, op := range r.Rules {
		if op.ID == id {
			r.Rules = append(r.Rules[:i], r.Rules[i+1:]...)
			return true
		}
	}
	return false
}

// RemoveByName 通过名称删除所有匹配的操作
func (r *Renamer) RemoveByName(name string) int {
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

// GetOperationByID 通过ID获取操作
func (r *Renamer) GetOperationByID(id string) *Rule {
	for _, op := range r.Rules {
		if op.ID == id {
			return &op
		}
	}
	return nil
}

// 单个文件名处理
func (r *Renamer) Apply(filename string) string {
	result := filename
	for _, op := range r.Rules {
		// 这里需要实现具体的重命名逻辑
		// 可以使用正则表达式等方式处理
		result = op.Apply(result)
	}
	return result
}

func (r *Renamer) ApplyBatchWithErrors(filenames []string) []RenameResult {
	results := make([]RenameResult, len(filenames))
	for i, filename := range filenames {
		results[i] = RenameResult{
			OldName: filename,
		}
		if filename == "" {
			results[i].Error = "空文件名"
			continue
		}
		newName := r.Apply(filename)
		if newName == "" {
			results[i].Error = "生成的文件名无效"
			continue
		}
		results[i].NewName = newName
	}
	return results
}

// SaveConfig 保存配置到JSON
func (r *Renamer) SaveConfig() ([]byte, error) {
	return json.Marshal(r)
}

// LoadConfig 从JSON加载配置
func (r *Renamer) LoadConfig(data []byte) error {
	return json.Unmarshal(data, r)
}
 