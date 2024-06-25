package identifier

import "errors"

var (
	// ErrTaskUnsupported indicates that a task is unsupported.
	ErrTaskUnsupported = errors.New("task unsupported")
)

// TaskIdentifier holds the identifier of a task.
type TaskIdentifier string
