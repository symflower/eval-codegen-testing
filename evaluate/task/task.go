package task

import (
	"fmt"

	pkgerrors "github.com/pkg/errors"
	"github.com/symflower/eval-dev-quality/language"
	"github.com/symflower/eval-dev-quality/log"
	"github.com/symflower/eval-dev-quality/model"
	evaltask "github.com/symflower/eval-dev-quality/task"
	"github.com/symflower/eval-dev-quality/task/identifier"
)

var (
	// AllIdentifiers holds all available task identifiers.
	AllIdentifiers []identifier.TaskIdentifier
	// LookupIdentifier holds a map of all available task identifiers.
	LookupIdentifier = map[identifier.TaskIdentifier]bool{}
)

// registerIdentifier registers the given identifier and makes it available.
func registerIdentifier(name string) (taskIdentifier identifier.TaskIdentifier) {
	taskIdentifier = identifier.TaskIdentifier(name)
	AllIdentifiers = append(AllIdentifiers, taskIdentifier)

	if _, ok := LookupIdentifier[taskIdentifier]; ok {
		panic(fmt.Sprintf("task identifier already registered: %s", taskIdentifier))
	}
	LookupIdentifier[taskIdentifier] = true

	return taskIdentifier
}

var (
	// IdentifierWriteTests holds the identifier for the "write test" task.
	IdentifierWriteTests = registerIdentifier("write-tests")
	// IdentifierCodeRepair holds the identifier for the "code repair" task.
	IdentifierCodeRepair = registerIdentifier("code-repair")
)

// TaskForIdentifier returns a task based on the task identifier.
func TaskForIdentifier(taskIdentifier identifier.TaskIdentifier, logger *log.Logger, resultPath string, model model.Model, language language.Language) (task evaltask.Task, err error) {
	switch taskIdentifier {
	case IdentifierWriteTests:
		return &TaskWriteTests{}, nil
	case IdentifierCodeRepair:
		return &TaskCodeRepair{}, nil
	default:
		return nil, pkgerrors.Wrap(identifier.ErrTaskUnsupported, string(taskIdentifier))
	}
}
