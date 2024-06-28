package task

import (
	"fmt"
	"path/filepath"

	pkgerrors "github.com/pkg/errors"
	"github.com/symflower/eval-dev-quality/language"
	"github.com/symflower/eval-dev-quality/log"
	evalmodel "github.com/symflower/eval-dev-quality/model"
	evaltask "github.com/symflower/eval-dev-quality/task"
)

var (
	// AllIdentifiers holds all available task identifiers.
	AllIdentifiers []evaltask.Identifier
	// LookupIdentifier holds a map of all available task identifiers.
	LookupIdentifier = map[evaltask.Identifier]bool{}
)

// registerIdentifier registers the given identifier and makes it available.
func registerIdentifier(name string) (identifier evaltask.Identifier) {
	identifier = evaltask.Identifier(name)
	AllIdentifiers = append(AllIdentifiers, identifier)

	if _, ok := LookupIdentifier[identifier]; ok {
		panic(fmt.Sprintf("task identifier already registered: %s", identifier))
	}
	LookupIdentifier[identifier] = true

	return identifier
}

var (
	// IdentifierWriteTests holds the identifier for the "write test" task.
	IdentifierWriteTests = registerIdentifier("write-tests")
	// IdentifierCodeRepair holds the identifier for the "code repair" task.
	IdentifierCodeRepair = registerIdentifier("code-repair")
)

// TaskForIdentifier returns a task based on the task identifier.
func TaskForIdentifier(taskIdentifier evaltask.Identifier, logger *log.Logger, resultPath string, model evalmodel.Model, language language.Language) (task evaltask.Task, err error) {
	switch taskIdentifier {
	case IdentifierWriteTests:
		return newTaskWriteTests(logger, resultPath, model, language), nil
	case IdentifierCodeRepair:
		return newCodeRepairTask(logger, resultPath, model, language), nil
	default:
		return nil, pkgerrors.Wrap(evaltask.ErrTaskUnsupported, string(taskIdentifier))
	}
}

// initializeLoggingForRun sets up logging for a task run.
func initializeLoggingForRun(parentLogger *log.Logger, resultPath string, task evaltask.Task, repository evaltask.Repository, model evalmodel.Model, language language.Language) (logger *log.Logger, finalizeRun func(problems []error), err error) {
	log, logClose, err := log.WithFile(
		parentLogger,
		filepath.Join(
			resultPath,
			string(task.Identifier()),
			evalmodel.CleanModelNameForFileSystem(model.ID()),
			language.ID(),
			repository.Name()+".log",
		),
	)
	if err != nil {
		return nil, nil, err
	}

	log.Printf("Evaluating model %q on task %q using language %q and repository %q", model.ID(), task.Identifier(), language.ID(), repository.Name())

	return parentLogger, func(problems []error) {
		log.Printf("Evaluated model %q on task %q using language %q and repository %q: encountered %d problems: %+v", model.ID(), task.Identifier(), language.ID(), repository.Name(), len(problems), problems)

		logClose()
	}, nil
}
