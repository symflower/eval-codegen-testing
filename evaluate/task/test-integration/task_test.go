package testintegration

import (
	"log"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/symflower/eval-dev-quality/evaluate/metrics"
	evaluatetask "github.com/symflower/eval-dev-quality/evaluate/task"
	tasktesting "github.com/symflower/eval-dev-quality/evaluate/task/testing"
	"github.com/symflower/eval-dev-quality/language/golang"
	"github.com/symflower/eval-dev-quality/model/symflower"
	"github.com/symflower/eval-dev-quality/task"
	evaltask "github.com/symflower/eval-dev-quality/task"
	"github.com/symflower/eval-dev-quality/tools"
	toolstesting "github.com/symflower/eval-dev-quality/tools/testing"
)

func TestTaskWriteTestsRun(t *testing.T) {
	toolstesting.RequiresTool(t, tools.NewSymflower())

	validate := func(t *testing.T, tc *tasktesting.TestCaseTask) {
		t.Run(tc.Name, func(t *testing.T) {
			tc.Validate(t,
				func(logger *log.Logger, testDataPath string, repositoryPathRelative string) (repository evaltask.Repository, cleanup func(), err error) {
					return evaluatetask.TemporaryRepository(logger, testDataPath, repositoryPathRelative)
				},
				func() (task task.Task, err error) {
					return evaluatetask.TaskForIdentifier(evaluatetask.IdentifierWriteTests)
				},
			)
		})
	}

	validate(t, &tasktesting.TestCaseTask{
		Name: "Plain",

		Model:          symflower.NewModel(),
		Language:       &golang.Language{},
		TestDataPath:   filepath.Join("..", "..", "..", "testdata"),
		RepositoryPath: filepath.Join("golang", "plain"),

		ExpectedRepositoryAssessment: metrics.Assessments{
			metrics.AssessmentKeyGenerateTestsForFileCharacterCount: 254,
			metrics.AssessmentKeyResponseCharacterCount:             254,
			metrics.AssessmentKeyCoverage:                           10,
			metrics.AssessmentKeyFilesExecuted:                      1,
			metrics.AssessmentKeyResponseNoError:                    1,
			metrics.AssessmentKeyResponseNoExcess:                   1,
			metrics.AssessmentKeyResponseWithCode:                   1,
		},
		ExpectedResultFiles: map[string]func(t *testing.T, filePath string, data string){
			filepath.Join(string(evaluatetask.IdentifierWriteTests), "symflower_symbolic-execution", "golang", "golang", "plain.log"): func(t *testing.T, filePath, data string) {
				assert.Contains(t, data, "Evaluating model \"symflower/symbolic-execution\"")
				assert.Contains(t, data, "Generated 1 test")
				assert.Contains(t, data, "PASS: TestSymflowerPlain")
				assert.Contains(t, data, "Evaluated model \"symflower/symbolic-execution\"")
			},
		},
	})
}
