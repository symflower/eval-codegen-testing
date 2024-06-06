package task

import (
	"errors"
	"log"

	"github.com/symflower/eval-dev-quality/language"
)

var (
	// ErrTaskUnsupported indicates that a task is unsupported.
	ErrTaskUnsupported = errors.New("task unsupported")
)

// Identifier holds the identifier of a task.
type Identifier string

// AllIdentifiers holds all available task identifiers.
var AllIdentifiers []Identifier

// registerIdentifier registers the given identifier and makes it available.
func registerIdentifier(name string) (identifier Identifier) {
	identifier = Identifier(name)
	AllIdentifiers = append(AllIdentifiers, identifier)

	return identifier
}

var (
	// IdentifierWriteTests holds the identifier for the "write test" task.
	IdentifierWriteTests = registerIdentifier("write-tests")
)

// Context holds the data needed for running a task.
type Context struct {
	// Language holds the language for which the task should be evaluated.
	Language language.Language

	// RepositoryPath holds the absolute path to the repository.
	RepositoryPath string
	// FilePath holds the path the file under test relative to the repository path.
	FilePath string

	// Logger is used for logging during evaluation.
	Logger *log.Logger
}
