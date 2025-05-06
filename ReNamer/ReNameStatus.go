package ReNamer

import (
	"encoding/json"
	"fmt"
)

// ReNameStatus represents the status of a rename operation
type ReNameStatus int

const (
	StatusPending ReNameStatus = iota // Not executed
	StatusSuccess                     // Execution successful
	StatusError                       // Execution failed
)

func (s ReNameStatus) String() string {
	switch s {
	case StatusPending:
		return "pending"
	case StatusSuccess:
		return "success"
	case StatusError:
		return "error"
	default:
		return "unknown"
	}
}

func (s *ReNameStatus) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}

	switch str {
	case "pending":
		*s = StatusPending
	case "success":
		*s = StatusSuccess
	case "error":
		*s = StatusError
	default:
		return fmt.Errorf("unknown status: %s", str)
	}
	return nil
}

// MarshalJSON implements the json.Marshaler interface
func (s ReNameStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}
