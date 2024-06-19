package task

import (
	"path/filepath"

	pkgerrors "github.com/pkg/errors"
	"github.com/symflower/eval-dev-quality/evaluate/metrics"
	"github.com/symflower/eval-dev-quality/log"
	"github.com/symflower/eval-dev-quality/model"
	evaltask "github.com/symflower/eval-dev-quality/task"
	"github.com/symflower/eval-dev-quality/task/identifier"
)

// TaskWriteTests holds the write test task.
type TaskWriteTests struct {
}

var _ evaltask.Task = (*TaskWriteTests)(nil)

// Identifier returns the write test task identifier.
func (t *TaskWriteTests) Identifier() identifier.TaskIdentifier {
	return IdentifierWriteTests
}

// TaskWriteTests generates test files for the given implementation file in a repository.
func (t *TaskWriteTests) Run(ctx evaltask.Context) (repositoryAssessment metrics.Assessments, problems []error, err error) {
	dataPath := ctx.Repository.DataPath()

	log, logClose, err := log.WithFile(ctx.Logger, filepath.Join(ctx.ResultPath, string(t.Identifier()), model.CleanModelNameForFileSystem(ctx.Model.ID()), ctx.Language.ID(), ctx.Repository.Name()+".log"))
	if err != nil {
		return nil, nil, err
	}
	defer logClose()

	log.Printf("Evaluating model %q on task %q using language %q and repository %q", ctx.Model.ID(), t.Identifier(), ctx.Language.ID(), ctx.Repository.Name())
	defer func() {
		log.Printf("Evaluated model %q on task %q using language %q and repository %q: encountered %d problems: %+v", ctx.Model.ID(), t.Identifier(), ctx.Language.ID(), ctx.Repository.Name(), len(problems), problems)
	}()

	filePaths, err := ctx.Language.Files(log, dataPath)
	if err != nil {
		return nil, problems, pkgerrors.WithStack(err)
	}

	repositoryAssessment = metrics.NewAssessments()
	for _, filePath := range filePaths {
		if err := ctx.Repository.Reset(ctx.Logger); err != nil {
			ctx.Logger.Panicf("ERROR: unable to reset temporary repository path: %s", err)
		}

		modelContext := model.Context{
			Language: ctx.Language,

			RepositoryPath: dataPath,
			FilePath:       filePath,

			Logger: log,
		}
		assessments, err := ctx.Model.RunTask(modelContext, t.Identifier())
		if err != nil {
			problems = append(problems, pkgerrors.WithMessage(err, filePath))

			continue
		}
		if assessments[metrics.AssessmentKeyProcessingTime] == 0 {
			return nil, nil, pkgerrors.Errorf("no model response time measurement present for %q at repository %q", ctx.Model.ID(), ctx.Repository.Name())
		}
		repositoryAssessment.Add(assessments)
		repositoryAssessment.Award(metrics.AssessmentKeyResponseNoError)

		coverage, ps, err := ctx.Language.Execute(log, dataPath)
		problems = append(problems, ps...)
		if err != nil {
			problems = append(problems, pkgerrors.WithMessage(err, filePath))

			continue
		}
		log.Printf("Executes tests with %d coverage objects", coverage)
		repositoryAssessment.Award(metrics.AssessmentKeyFilesExecuted)
		repositoryAssessment.AwardPoints(metrics.AssessmentKeyCoverage, coverage)
	}

	return repositoryAssessment, problems, nil
}
