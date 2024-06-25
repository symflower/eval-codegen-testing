package model

import (
	"log"

	"github.com/symflower/eval-dev-quality/evaluate/metrics"
	"github.com/symflower/eval-dev-quality/language"
	"github.com/symflower/eval-dev-quality/task/identifier"
)

// Model defines a model that can be queried for generations.
type Model interface {
	// ID returns the unique ID of this model.
	ID() (id string)

	// IsTaskSupported returns whether the model supports the given task or not.
	IsTaskSupported(taskIdentifier identifier.TaskIdentifier) (isSupported bool)
	// RunTask runs the given task.
	RunTask(ctx Context, taskIdentifier identifier.TaskIdentifier) (assessments metrics.Assessments, err error)
}

// Context holds the data needed by a model for running a task.
type Context struct {
	// Language holds the language for which the task should be evaluated.
	Language language.Language

	// RepositoryPath holds the absolute path to the repository.
	RepositoryPath string
	// FilePath holds the path the file under test relative to the repository path.
	FilePath string

	// Arguments holds extra data that can be used in a query prompt.
	Arguments any

	// Logger is used for logging during evaluation.
	Logger *log.Logger
}

// SetQueryAttempts defines a model that can set the number of query attempts when a model request errors in the process of solving a task.
type SetQueryAttempts interface {
	// SetQueryAttempts sets the number of query attempts to perform when a model request errors in the process of solving a task.
	SetQueryAttempts(attempts uint)
}
