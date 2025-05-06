package renamer

type ReNameMode int

const (
	ModeNormal ReNameMode = iota // 普通模式，跳过错误状态的映射
	ModeError                    // 错误重试模式，只重试错误状态的映射
	ModeUndo                     // 回退模式，执行映射回退
)

func (m ReNameMode) String() string {
	switch m {
	case ModeNormal:
		return "normal"
	case ModeError:
		return "error"
	case ModeUndo:
		return "undo"
	default:
		return "unknown"
	}
}
