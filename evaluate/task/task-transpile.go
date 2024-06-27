package task

import (
	"os"
	"path/filepath"
	"strings"

	pkgerrors "github.com/pkg/errors"
	"github.com/symflower/eval-dev-quality/evaluate/metrics"
	"github.com/symflower/eval-dev-quality/language"
	"github.com/symflower/eval-dev-quality/language/golang"
	"github.com/symflower/eval-dev-quality/language/java"
	"github.com/symflower/eval-dev-quality/log"
	"github.com/symflower/eval-dev-quality/model"
	evaltask "github.com/symflower/eval-dev-quality/task"
)

// TaskTranspile holds the transpilation task.
type TaskTranspile struct {
	// ResultPath holds the directory path where results should be written to.
	ResultPath string

	// Language holds the language for which the task should be evaluated.
	Language language.Language
	// Model holds the model which the task should be evaluated.
	Model model.Model

	// Logger holds the logger for this tasks.
	Logger *log.Logger
}

// TaskArgumentsTranspile holds extra arguments to be used in a query prompt.
type TaskArgumentsTranspile struct {
	// TargetLanguage holds the target language for transpilation.
	TargetLanguage language.Language
	// StubFilePath holds the path for the file containing just the function signature of the language we are transpiling to.
	StubFilePath string
}

var _ evaltask.Task = (*TaskTranspile)(nil)

// newTaskTranspile returns a transpilation task.
func newTaskTranspile(logger *log.Logger, resultPath string, model model.Model, language language.Language) (task evaltask.Task) {
	return &TaskTranspile{
		ResultPath: resultPath,
		Language:   language,
		Model:      model,
		Logger:     logger,
	}
}

// Identifier returns the transpilation task identifier.
func (t *TaskTranspile) Identifier() evaltask.Identifier {
	return IdentifierTranspile
}

// Run transpiles code between Go and Java and runs predefined tests to check if the transpilation was successful.
func (t *TaskTranspile) Run(repository evaltask.Repository) (repositoryAssessment metrics.Assessments, problems []error, err error) {
	log, logClose, err := log.WithFile(t.Logger, filepath.Join(t.ResultPath, string(t.Identifier()), model.CleanModelNameForFileSystem(t.Model.ID()), t.Language.ID(), repository.Name()+".log"))
	if err != nil {
		return nil, nil, err
	}
	defer logClose()

	log.Printf("Evaluating model %q on task %q using language %q and repository %q", t.Model.ID(), t.Identifier(), t.Language.ID(), repository.Name())
	defer func() {
		log.Printf("Evaluated model %q on task %q using language %q and repository %q: encountered %d problems: %+v", t.Model.ID(), t.Identifier(), t.Language.ID(), repository.Name(), len(problems), problems)
	}()

	var packagePaths []string
	files, err := os.ReadDir(repository.DataPath())
	if err != nil {
		return nil, nil, pkgerrors.WithStack(err)
	}
	for _, file := range files {
		if file.IsDir() && !strings.HasPrefix(file.Name(), ".") { // Ignore hidden directories.
			packagePaths = append(packagePaths, filepath.Join(repository.DataPath(), file.Name()))
		}
	}

	repositoryAssessment = metrics.NewAssessments()
	for _, packagePath := range packagePaths {
		if err := repository.Reset(t.Logger); err != nil {
			t.Logger.Panicf("ERROR: unable to reset temporary repository path: %s", err)
		}

		var targetLanguage language.Language
		if _, ok := t.Language.(*golang.Language); ok {
			targetLanguage = &java.Language{}
		} else {
			targetLanguage = &golang.Language{}
		}

		sourceFilePath, stubFilePath, err := t.unpackTranspilerPackage(log, repository, targetLanguage, packagePath)
		if err != nil {
			return nil, nil, err
		}

		ctx := evaltask.Context{
			Language: t.Language,

			RepositoryPath: packagePath,
			FilePath:       sourceFilePath,

			Arguments: &TaskArgumentsTranspile{
				TargetLanguage: targetLanguage,
				StubFilePath:   stubFilePath,
			},

			Logger: log,
		}
		assessments, err := t.Model.RunTask(ctx, t.Identifier())
		if err != nil {
			problems = append(problems, pkgerrors.WithMessage(err, sourceFilePath))

			continue
		}
		if assessments[metrics.AssessmentKeyProcessingTime] == 0 {
			return nil, nil, pkgerrors.Errorf("no model response time measurement present for %q at repository %q", t.Model.ID(), repository.Name())
		}
		repositoryAssessment.Add(assessments)
		repositoryAssessment.Award(metrics.AssessmentKeyResponseNoError)

		coverage, ps, err := targetLanguage.Execute(log, packagePath)
		problems = append(problems, ps...)
		if err != nil {
			problems = append(problems, pkgerrors.WithMessage(err, sourceFilePath))

			continue
		}
		log.Printf("Executes tests with %d coverage objects", coverage)
		repositoryAssessment.Award(metrics.AssessmentKeyFilesExecuted)
		repositoryAssessment.AwardPoints(metrics.AssessmentKeyCoverage, coverage)
	}

	return repositoryAssessment, problems, nil
}

// unpackTranspilerPackage checks if the testdata repository for the transpilation task is well-formed and returns the path to the implementation file and also the path to the file that holds the stub.
func (t *TaskTranspile) unpackTranspilerPackage(fileLogger *log.Logger, repository evaltask.Repository, targetLanguage language.Language, packagePath string) (sourceFilePath string, stubFilePath string, err error) {
	// Check if the package path has a directory called "implementation".
	var implementationPath string
	packageFiles, err := os.ReadDir(packagePath)
	if err != nil {
		return "", "", pkgerrors.WithStack(err)
	}
	for _, file := range packageFiles {
		if file.IsDir() && file.Name() == "implementation" {
			implementationPath = filepath.Join(packagePath, file.Name())
		}
	}
	if implementationPath == "" {
		return "", "", pkgerrors.Errorf("package %q in repository %q must have an \"implementation\" directory with a %s source file to transpile", packagePath, repository.Name(), t.Language.Name())
	}

	// Check if the "implementation" directory contains only one source file in the in the language being transpiled from..
	implementationFiles, err := os.ReadDir(implementationPath)
	if err != nil {
		return "", "", pkgerrors.WithStack(err)
	} else if len(implementationFiles) != 1 {
		return "", "", pkgerrors.Errorf("package %q in repository %q must have an \"implementation\" directory with just one %s source file to transpile", packagePath, repository.Name(), t.Language.Name())
	}
	sourceFilePath = filepath.Join(implementationPath, implementationFiles[0].Name())
	if filepath.Ext(sourceFilePath) != t.Language.DefaultFileExtension() {
		return "", "", pkgerrors.Errorf("package %q in repository %q must have an \"implementation\" directory with a %s source file", packagePath, repository.Name(), t.Language.Name())
	} else if strings.HasPrefix(sourceFilePath, t.Language.DefaultTestFileSuffix()) {
		return "", "", pkgerrors.Errorf("package %q in repository %q must have an \"implementation\" directory with only a %s source file, but found a test file %q", packagePath, repository.Name(), t.Language.Name(), sourceFilePath)
	}

	// Check if the package path contains a source file and the corresponding test file in the target language for transpilation.
	var hasTestFile bool
	files, err := targetLanguage.Files(fileLogger, packagePath)
	if err != nil {
		return "", "", pkgerrors.WithStack(err)
	} else if len(files) != 2 {
		return "", "", pkgerrors.Errorf("package %q in repository %q must have a %s source file and the corresponding test file", packagePath, repository.Name(), targetLanguage.Name())
	}
	for _, file := range files {
		if strings.HasSuffix(file, targetLanguage.DefaultTestFileSuffix()) {
			hasTestFile = true
		} else if filepath.Ext(file) == targetLanguage.DefaultFileExtension() {
			stubFilePath = filepath.Join(packagePath, file)
		}
	}
	if !hasTestFile {
		return "", "", pkgerrors.Errorf("package %q in repository %q must have a %s test file", packagePath, repository.Name(), targetLanguage.Name())
	} else if stubFilePath == "" {
		return "", "", pkgerrors.Errorf("package %q in repository %q must have a %s file with a stub function", packagePath, repository.Name(), targetLanguage.Name())
	}

	return sourceFilePath, stubFilePath, nil
}
