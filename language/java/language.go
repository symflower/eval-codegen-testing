package java

import (
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	pkgerrors "github.com/pkg/errors"
	"github.com/zimmski/osutil"

	"github.com/symflower/eval-dev-quality/language"
	"github.com/symflower/eval-dev-quality/log"
	"github.com/symflower/eval-dev-quality/tools"
	"github.com/symflower/eval-dev-quality/util"
)

// Language holds a Java language to evaluate a repository.
type Language struct{}

func init() {
	language.Register(&Language{})
}

var _ language.Language = (*Language)(nil)

// ID returns the unique ID of this language.
func (l *Language) ID() (id string) {
	return "java"
}

// Name is the prose name of this language.
func (l *Language) Name() (id string) {
	return "Java"
}

// Files returns a list of relative file paths of the repository that should be evaluated.
func (l *Language) Files(logger *log.Logger, repositoryPath string) (filePaths []string, err error) {
	repositoryPath, err = filepath.Abs(repositoryPath)
	if err != nil {
		return nil, pkgerrors.WithStack(err)
	}

	fs, err := osutil.FilesRecursive(repositoryPath)
	if err != nil {
		return nil, pkgerrors.WithStack(err)
	}

	repositoryPath = repositoryPath + string(os.PathSeparator)
	for _, f := range fs {
		if !strings.HasSuffix(f, ".java") {
			continue
		}

		filePaths = append(filePaths, strings.TrimPrefix(f, repositoryPath))
	}

	return filePaths, nil
}

// ImportPath returns the import path of the given source file.
func (l *Language) ImportPath(projectRootPath string, filePath string) (importPath string) {
	importPath = strings.ReplaceAll(filepath.Dir(filePath), string(os.PathSeparator), ".")

	t := "src.main.java"
	if l := strings.LastIndex(importPath, t); l != -1 {
		return importPath[l+len(t)+1:]
	}

	return importPath
}

// TestFilePath returns the file path of a test file given the corresponding file path of the test's source file.
func (l *Language) TestFilePath(projectRootPath string, filePath string) (testFilePath string) {
	if l := strings.LastIndex(filePath, "src/main/java"); l != -1 {
		t := "src/test/java"
		filePath = filePath[:l] + t + filePath[l+len(t):]
	}

	return strings.TrimSuffix(filePath, ".java") + "Test.java"
}

// TestFramework returns the human-readable name of the test framework that should be used.
func (l *Language) TestFramework() (testFramework string) {
	return "JUnit 5"
}

var languageJavaCoverageMatch = regexp.MustCompile(`Total coverage (.+?)%`)

// Execute invokes the language specific testing on the given repository.
func (l *Language) Execute(logger *log.Logger, repositoryPath string) (coverage float64, err error) {
	commandOutput, err := util.CommandWithResult(logger, &util.Command{
		Command: []string{
			tools.SymflowerPath, "test",
			"--language", "java",
			"--workspace", repositoryPath,
		},

		Directory: repositoryPath,
	})
	if err != nil {
		return 0.0, pkgerrors.WithStack(err)
	}

	mc := languageJavaCoverageMatch.FindStringSubmatch(commandOutput)
	if mc == nil {
		return 0.0, pkgerrors.WithStack(pkgerrors.WithMessage(errors.New("could not find coverage report"), commandOutput))
	}
	coverage, err = strconv.ParseFloat(mc[1], 64)
	if err != nil {
		return 0.0, pkgerrors.WithStack(err)
	}

	return coverage, nil
}
