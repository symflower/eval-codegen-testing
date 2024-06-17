package llm

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/avast/retry-go"
	pkgerrors "github.com/pkg/errors"
	"github.com/zimmski/osutil/bytesutil"

	"github.com/symflower/eval-dev-quality/evaluate/metrics"
	evaluatetask "github.com/symflower/eval-dev-quality/evaluate/task"
	"github.com/symflower/eval-dev-quality/language"
	"github.com/symflower/eval-dev-quality/model"
	"github.com/symflower/eval-dev-quality/model/llm/prompt"
	"github.com/symflower/eval-dev-quality/provider"
	"github.com/symflower/eval-dev-quality/task"
)

// Model represents a LLM model accessed via a provider.
type Model struct {
	// provider is the client to query the LLM model.
	provider provider.Query
	// model holds the identifier for the LLM model.
	model string

	// queryAttempts holds the number of query attempts to perform when a model request errors in the process of solving a task.
	queryAttempts uint
}

// NewModel returns an LLM model corresponding to the given identifier which is queried via the given provider.
func NewModel(provider provider.Query, modelIdentifier string) *Model {
	return &Model{
		provider: provider,
		model:    modelIdentifier,

		queryAttempts: 1,
	}
}

// llmSourceFilePromptContext is the context for template for generating an LLM test generation prompt.
type llmSourceFilePromptContext struct {
	// Language holds the programming language name.
	Language language.Language

	// Code holds the source code of the file.
	Code string
	// FilePath holds the file path of the file.
	FilePath string
	// ImportPath holds the import path of the file.
	ImportPath string
}

// llmGenerateTestForFilePromptTemplate is the template for generating an LLM test generation prompt.
var llmGenerateTestForFilePromptTemplate = template.Must(template.New("model-llm-generate-test-for-file-prompt").Parse(bytesutil.StringTrimIndentations(`
	Given the following {{ .Language.Name }} code file "{{ .FilePath }}" with package "{{ .ImportPath }}", provide a test file for this code{{ with $testFramework := .Language.TestFramework }} with {{ $testFramework }} as a test framework{{ end }}.
	The tests should produce 100 percent code coverage and must compile.
	The response must contain only the test code and nothing else.

	` + "```" + `{{ .Language.ID }}
	{{ .Code }}
	` + "```" + `
`)))

// llmGenerateTestForFilePrompt returns the prompt for generating an LLM test generation.
func llmGenerateTestForFilePrompt(data *llmSourceFilePromptContext) (message string, err error) {
	data.Code = strings.TrimSpace(data.Code)

	var b strings.Builder
	if err := llmGenerateTestForFilePromptTemplate.Execute(&b, data); err != nil {
		return "", pkgerrors.WithStack(err)
	}

	return b.String(), nil
}

var _ model.Model = (*Model)(nil)

// ID returns the unique ID of this model.
func (m *Model) ID() (id string) {
	return m.model
}

// IsTaskSupported returns whether the model supports the given task or not.
func (m *Model) IsTaskSupported(taskIdentifier task.Identifier) (isSupported bool) {
	switch taskIdentifier {
	case evaluatetask.IdentifierWriteTests:
		return true
	case evaluatetask.IdentifierCodeRepair:
		return false
	default:
		return false
	}
}

// RunTask runs the given task.
func (m *Model) RunTask(ctx task.Context, taskIdentifier task.Identifier) (assessments metrics.Assessments, err error) {
	switch taskIdentifier {
	case evaluatetask.IdentifierWriteTests:
		return m.generateTestsForFile(ctx)
	default:
		return nil, pkgerrors.Wrap(task.ErrTaskUnsupported, string(taskIdentifier))
	}
}

// generateTestsForFile generates test files for the given implementation file in a repository.
func (m *Model) generateTestsForFile(ctx task.Context) (assessment metrics.Assessments, err error) {
	data, err := os.ReadFile(filepath.Join(ctx.RepositoryPath, ctx.FilePath))
	if err != nil {
		return nil, pkgerrors.WithStack(err)
	}
	fileContent := strings.TrimSpace(string(data))

	importPath := ctx.Language.ImportPath(ctx.RepositoryPath, ctx.FilePath)

	request, err := llmGenerateTestForFilePrompt(&llmSourceFilePromptContext{
		Language: ctx.Language,

		Code:       fileContent,
		FilePath:   ctx.FilePath,
		ImportPath: importPath,
	})
	if err != nil {
		return nil, err
	}

	response, duration, err := m.query(ctx.Logger, request)
	if err != nil {
		return nil, pkgerrors.WithStack(err)
	}

	assessment, testContent, err := prompt.ParseResponse(response)
	if err != nil {
		return nil, pkgerrors.WithStack(err)
	}
	assessment[metrics.AssessmentKeyProcessingTime] = uint64(duration.Milliseconds())
	assessment[metrics.AssessmentKeyResponseCharacterCount] = uint64(len(response))
	assessment[metrics.AssessmentKeyGenerateTestsForFileCharacterCount] = uint64(len(testContent))

	if err != nil {
		return nil, pkgerrors.WithStack(err)
	}

	testFilePath := ctx.Language.TestFilePath(ctx.RepositoryPath, ctx.FilePath)
	if err := os.MkdirAll(filepath.Join(ctx.RepositoryPath, filepath.Dir(testFilePath)), 0755); err != nil {
		return nil, pkgerrors.WithStack(err)
	}
	if err := os.WriteFile(filepath.Join(ctx.RepositoryPath, testFilePath), []byte(testContent), 0644); err != nil {
		return nil, pkgerrors.WithStack(err)
	}

	return assessment, nil
}

func (m *Model) query(log *log.Logger, request string) (response string, duration time.Duration, err error) {
	if err := retry.Do(
		func() error {
			log.Printf("Querying model %q with:\n%s", m.ID(), string(bytesutil.PrefixLines([]byte(request), []byte("\t"))))
			start := time.Now()
			response, err = m.provider.Query(context.Background(), m.model, request)
			if err != nil {
				return err
			}
			duration = time.Since(start)
			log.Printf("Model %q responded (%d ms) with:\n%s", m.ID(), duration.Milliseconds(), string(bytesutil.PrefixLines([]byte(response), []byte("\t"))))

			return nil
		},
		retry.Attempts(m.queryAttempts),
		retry.Delay(5*time.Second),
		retry.DelayType(retry.BackOffDelay),
		retry.LastErrorOnly(true),
		retry.OnRetry(func(n uint, err error) {
			log.Printf("Attempt %d/%d: %s", n+1, m.queryAttempts, err)
		}),
	); err != nil {
		return "", 0, err
	}

	return response, duration, nil
}

var _ model.SetQueryAttempts = (*Model)(nil)

// SetQueryAttempts sets the number of query attempts to perform when a model request errors in the process of solving a task.
func (m *Model) SetQueryAttempts(queryAttempts uint) {
	m.queryAttempts = queryAttempts
}
