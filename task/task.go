package task

import (
	"log"

	"github.com/symflower/eval-dev-quality/evaluate/metrics"
	"github.com/symflower/eval-dev-quality/language"
	"github.com/symflower/eval-dev-quality/model"
	"github.com/symflower/eval-dev-quality/task/identifier"
)

// Context holds the data need by a task to be run.
type Context struct {
	// Language holds the language for which the task should be evaluated.
	Language language.Language
	// Repository holds the repository which should be evaluated.
	Repository Repository
	// Model holds the model which the task should be evaluated.
	Model model.Model

	// ResultPath holds the directory path where results should be written to.
	ResultPath string

	// Logger holds the logger for this tasks.
	Logger *log.Logger
}

// Task defines an evaluation task.
type Task interface {
	// Identifier returns the task identifier.
	Identifier() (identifier identifier.TaskIdentifier)

	// Run runs a task in a given repository.
	Run(ctx Context) (assessments metrics.Assessments, problems []error, err error)
}

// Repository defines a repository to be evaluated.
type Repository interface {
	// Name holds the name of the repository.
	Name() (name string)
	// DataPath holds the absolute path to the repository.
	DataPath() (dataPath string)

	// SupportedTasks returns the list of task identifiers the repository supports.
	SupportedTasks() (tasks []identifier.TaskIdentifier)

	// Reset resets the repository to its initial state.
	Reset(logger *log.Logger) (err error)
}
