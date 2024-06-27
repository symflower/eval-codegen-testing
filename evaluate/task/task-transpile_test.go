package task

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/symflower/eval-dev-quality/evaluate/metrics"
	metricstesting "github.com/symflower/eval-dev-quality/evaluate/metrics/testing"
	tasktesting "github.com/symflower/eval-dev-quality/evaluate/task/testing"
	"github.com/symflower/eval-dev-quality/language/golang"
	"github.com/symflower/eval-dev-quality/language/java"
	"github.com/symflower/eval-dev-quality/log"
	modeltesting "github.com/symflower/eval-dev-quality/model/testing"
	"github.com/zimmski/osutil"
	"github.com/zimmski/osutil/bytesutil"
)

func TestTaskTranspileRun(t *testing.T) {
	validate := func(t *testing.T, tc *tasktesting.TestCaseTask) {
		t.Run(tc.Name, func(t *testing.T) {
			resultPath := t.TempDir()

			logOutput, logger := log.Buffer()
			defer func() {
				if t.Failed() {
					t.Logf("Logging output: %s", logOutput.String())
				}
			}()
			repository, cleanup, err := TemporaryRepository(logger, tc.TestDataPath, tc.RepositoryPath)
			assert.NoError(t, err)
			defer cleanup()

			taskTranspile := newTaskTranspile(logger, resultPath, tc.Model, tc.Language)
			tc.Validate(t, taskTranspile, repository, resultPath)
		})
	}

	t.Run("Transpile Go into Java", func(t *testing.T) {
		{
			temporaryDirectoryPath := t.TempDir()

			repositoryPath := filepath.Join(temporaryDirectoryPath, "golang", "transpile", "cascadingIfElse")
			require.NoError(t, osutil.CopyTree(filepath.Join("..", "..", "testdata", "golang", "transpile", "cascadingIfElse"), repositoryPath))

			modelMock := modeltesting.NewMockModelNamed(t, "mocked-model")

			transpiledSourceFilePath := filepath.Join("src", "main", "java", "com", "eval", "CascadingIfElse.java")
			transpiledSourceFileContent := bytesutil.StringTrimIndentations(`
				package com.eval;

				class CascadingIfElse {
					static int cascadingIfElse(int i) {
						if (i == 1) {
							return 2;
						} else if (i == 3) {
							return 4;
						} else {
							return 5;
						}
					}
				}
			`)
			modelMock.RegisterGenerateSuccess(t, IdentifierTranspile, transpiledSourceFilePath, transpiledSourceFileContent, metricstesting.AssessmentsWithProcessingTime).Once()

			validate(t, &tasktesting.TestCaseTask{
				Name: "Single test case",

				Model:          modelMock,
				Language:       &golang.Language{},
				TestDataPath:   temporaryDirectoryPath,
				RepositoryPath: filepath.Join("golang", "transpile"),

				ExpectedRepositoryAssessment: metrics.Assessments{
					metrics.AssessmentKeyCoverage:        80,
					metrics.AssessmentKeyFilesExecuted:   1,
					metrics.AssessmentKeyResponseNoError: 1,
				},
				ExpectedResultFiles: map[string]func(t *testing.T, filePath string, data string){
					filepath.Join(string(IdentifierTranspile), "mocked-model", "golang", "golang", "transpile.log"): func(t *testing.T, filePath, data string) {
						assert.Contains(t, data, "BUILD SUCCESS")
					},
				},
			})
		}
		{
			temporaryDirectoryPath := t.TempDir()

			repositoryPath := filepath.Join(temporaryDirectoryPath, "golang", "transpile")
			require.NoError(t, osutil.CopyTree(filepath.Join("..", "..", "testdata", "golang", "transpile", "cascadingIfElse"), filepath.Join(repositoryPath, "cascadingIfElse")))
			require.NoError(t, osutil.CopyTree(filepath.Join("..", "..", "testdata", "golang", "transpile", "sort"), filepath.Join(repositoryPath, "sort")))

			modelMock := modeltesting.NewMockModelNamed(t, "mocked-model")

			transpiledSourceFilePath := filepath.Join("src", "main", "java", "com", "eval", "CascadingIfElse.java")
			transpiledSourceFileContent := bytesutil.StringTrimIndentations(`
				package com.eval;

				class CascadingIfElse {
					static int cascadingIfElse(int i) {
						if (i == 1) {
							return 2;
						} else if (i == 3) {
							return 4;
						} else {
							return 5;
						}
					}
				}
			`)
			modelMock.RegisterGenerateSuccess(t, IdentifierTranspile, transpiledSourceFilePath, transpiledSourceFileContent, metricstesting.AssessmentsWithProcessingTime).Once()

			transpiledSourceFilePath = filepath.Join("src", "main", "java", "com", "eval", "Sort.java")
			transpiledSourceFileContent = bytesutil.StringTrimIndentations(`
				package com.eval;

				class Sort {
					static boolean isSorted(int[] a) {
						int i = 0;
						while (i < a.length - 1 && a[i] <= a[i + 1]) {
							i++;
						}

						return i == a.length - 1;
					}
				}
			`)
			modelMock.RegisterGenerateSuccess(t, IdentifierTranspile, transpiledSourceFilePath, transpiledSourceFileContent, metricstesting.AssessmentsWithProcessingTime).Once()

			validate(t, &tasktesting.TestCaseTask{
				Name: "Multiple test cases",

				Model:          modelMock,
				Language:       &golang.Language{},
				TestDataPath:   temporaryDirectoryPath,
				RepositoryPath: filepath.Join("golang", "transpile"),

				ExpectedRepositoryAssessment: metrics.Assessments{
					metrics.AssessmentKeyCoverage:        140,
					metrics.AssessmentKeyFilesExecuted:   2,
					metrics.AssessmentKeyResponseNoError: 2,
				},
				ExpectedResultFiles: map[string]func(t *testing.T, filePath string, data string){
					filepath.Join(string(IdentifierTranspile), "mocked-model", "golang", "golang", "transpile.log"): func(t *testing.T, filePath, data string) {
						assert.Contains(t, data, "BUILD SUCCESS")
					},
				},
			})
		}
	})
	t.Run("Transpile Java into Go", func(t *testing.T) {
		{
			temporaryDirectoryPath := t.TempDir()

			repositoryPath := filepath.Join(temporaryDirectoryPath, "java", "transpile", "cascadingIfElse")
			require.NoError(t, osutil.CopyTree(filepath.Join("..", "..", "testdata", "java", "transpile", "cascadingIfElse"), repositoryPath))

			modelMock := modeltesting.NewMockModelNamed(t, "mocked-model")

			transpiledSourceFilePath := filepath.Join("cascadingIfElse.go")
			transpiledSourceFileContent := bytesutil.StringTrimIndentations(`
				package cascadingIfElse

				func cascadingIfElse(i int) int {
					if i == 1 {
						return 2
					} else if i == 3 {
						return 4
					} else {
						return 5
					}
				}
			`)
			modelMock.RegisterGenerateSuccess(t, IdentifierTranspile, transpiledSourceFilePath, transpiledSourceFileContent, metricstesting.AssessmentsWithProcessingTime).Once()

			validate(t, &tasktesting.TestCaseTask{
				Name: "Single test case",

				Model:          modelMock,
				Language:       &java.Language{},
				TestDataPath:   temporaryDirectoryPath,
				RepositoryPath: filepath.Join("java", "transpile"),

				ExpectedRepositoryAssessment: metrics.Assessments{
					metrics.AssessmentKeyCoverage:        30,
					metrics.AssessmentKeyFilesExecuted:   1,
					metrics.AssessmentKeyResponseNoError: 1,
				},
				ExpectedResultFiles: map[string]func(t *testing.T, filePath string, data string){
					filepath.Join(string(IdentifierTranspile), "mocked-model", "java", "java", "transpile.log"): func(t *testing.T, filePath, data string) {
						assert.Contains(t, data, "PASS: TestSymflowerCascadingIfElse/#00")
						assert.Contains(t, data, "PASS: TestSymflowerCascadingIfElse/#01")
						assert.Contains(t, data, "PASS: TestSymflowerCascadingIfElse/#02")
					},
				},
			})
		}
		{
			temporaryDirectoryPath := t.TempDir()

			repositoryPath := filepath.Join(temporaryDirectoryPath, "java", "transpile")
			require.NoError(t, osutil.CopyTree(filepath.Join("..", "..", "testdata", "java", "transpile", "cascadingIfElse"), filepath.Join(repositoryPath, "cascadingIfElse")))
			require.NoError(t, osutil.CopyTree(filepath.Join("..", "..", "testdata", "java", "transpile", "sort"), filepath.Join(repositoryPath, "sort")))

			modelMock := modeltesting.NewMockModelNamed(t, "mocked-model")

			transpiledSourceFilePath := filepath.Join("cascadingIfElse.go")
			transpiledSourceFileContent := bytesutil.StringTrimIndentations(`
				package cascadingIfElse

				func cascadingIfElse(i int) int {
					if i == 1 {
						return 2
					} else if i == 3 {
						return 4
					} else {
						return 5
					}
				}
			`)
			modelMock.RegisterGenerateSuccess(t, IdentifierTranspile, transpiledSourceFilePath, transpiledSourceFileContent, metricstesting.AssessmentsWithProcessingTime).Once()

			transpiledSourceFilePath = filepath.Join("sort.go")
			transpiledSourceFileContent = bytesutil.StringTrimIndentations(`
				package isSorted

				func isSorted(a []int) bool {
					i := 0
					for i < len(a)-1 && a[i] <= a[i+1] {
						i++
					}

					return i == len(a)-1
				}
			`)
			modelMock.RegisterGenerateSuccess(t, IdentifierTranspile, transpiledSourceFilePath, transpiledSourceFileContent, metricstesting.AssessmentsWithProcessingTime).Once()

			validate(t, &tasktesting.TestCaseTask{
				Name: "Multiple test cases",

				Model:          modelMock,
				Language:       &java.Language{},
				TestDataPath:   temporaryDirectoryPath,
				RepositoryPath: filepath.Join("java", "transpile"),

				ExpectedRepositoryAssessment: metrics.Assessments{
					metrics.AssessmentKeyCoverage:        50,
					metrics.AssessmentKeyFilesExecuted:   2,
					metrics.AssessmentKeyResponseNoError: 2,
				},
				ExpectedResultFiles: map[string]func(t *testing.T, filePath string, data string){
					filepath.Join(string(IdentifierTranspile), "mocked-model", "java", "java", "transpile.log"): func(t *testing.T, filePath, data string) {
						assert.Contains(t, data, "PASS: TestSymflowerCascadingIfElse/#00")
						assert.Contains(t, data, "PASS: TestSymflowerCascadingIfElse/#01")
						assert.Contains(t, data, "PASS: TestSymflowerCascadingIfElse/#02")

						assert.Contains(t, data, "PASS: TestSymflowerIsSorted/#00")
						assert.Contains(t, data, "PASS: TestSymflowerIsSorted/#01")
						assert.Contains(t, data, "PASS: TestSymflowerIsSorted/#02")
						assert.Contains(t, data, "PASS: TestSymflowerIsSorted/#03")
						assert.Contains(t, data, "PASS: TestSymflowerIsSorted/#04")
					},
				},
			})
		}
	})
}
