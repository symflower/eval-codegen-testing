package report

import (
	"io"
	"os"
	"path/filepath"
	"text/template"
	"time"

	pkgerrors "github.com/pkg/errors"
	"github.com/zimmski/osutil/bytesutil"

	"github.com/symflower/eval-dev-quality/evaluate/metrics"
)

// Markdown holds the values for exporting a Markdown report.
type Markdown struct {
	// DateTime holds the timestamp of the evaluation.
	DateTime time.Time
	// Version holds the version of the evaluation tool.
	Version string

	// CSVPath holds the path of detailed CSV results.
	CSVPath string
	// LogPath holds the path of detailed logs.
	LogPath string

	// AssessmentPerModel holds
	AssessmentPerModel map[string]metrics.Assessments
	// TotalScore holds the total reachable score per task.
	// REMARK Used for category computation.
	TotalScore uint
}

// markdownTemplateContext holds the template for a Markdown report.
type markdownTemplateContext struct {
	Markdown

	Categories        []*metrics.AssessmentCategory
	ModelsPerCategory map[*metrics.AssessmentCategory][]string
}

// markdownTemplate holds the template for a Markdown report.
var markdownTemplate = template.Must(template.New("template-report").Parse(bytesutil.StringTrimIndentations(`
	# Evaluation from {{.DateTime.Format "2006-01-02 15:04:05"}}

	This report was generated by [DevQualityEval benchmark](https://github.com/symflower/eval-dev-quality) in ` + "`" + `version {{.Version}}` + "`" + `.

	## Results

	> Keep in mind that LLMs are nondeterministic. The following results just reflect a current snapshot.

	The results of all models have been divided into the following categories:
	{{ range $category := .Categories -}}
	- {{ $category.Name }}: {{ $category.Description }}
	{{ end }}
	The following sections list all models with their categories. The complete log of the evaluation with all outputs can be found [here]({{.LogPath}}). Detailed scoring can be found [here]({{.CSVPath}}).

	{{ range $category := .Categories -}}
	{{ with $modelNames := index $.ModelsPerCategory $category -}}
	### "{{ $category.Name }}"

	{{ $category.Description }}

	{{ range $modelName := $modelNames -}}
	- ` + "`" + `{{ $modelName }}` + "`" + `
	{{ end }}
	{{ end }}
	{{- end -}}
`)))

// Format formats the markdown values in the template to the given writer.
func (m Markdown) Format(writer io.Writer) error {
	templateContext := markdownTemplateContext{
		Markdown:   m,
		Categories: metrics.AllAssessmentCategories,
	}
	templateContext.ModelsPerCategory = make(map[*metrics.AssessmentCategory][]string, len(metrics.AllAssessmentCategories))
	for model, assessment := range m.AssessmentPerModel {
		category := assessment.Category(m.TotalScore)
		templateContext.ModelsPerCategory[category] = append(templateContext.ModelsPerCategory[category], model)
	}
	// TODO Generate svg using maybe https://github.com/wcharczuk/go-chart.

	return pkgerrors.WithStack(markdownTemplate.Execute(writer, templateContext))
}

// WriteToFile writes the Markdown values in the template to the given file.
func (t Markdown) WriteToFile(path string) (err error) {
	t.CSVPath, err = filepath.Abs(t.CSVPath)
	if err != nil {
		return err
	}
	t.LogPath, err = filepath.Abs(t.LogPath)
	if err != nil {
		return err
	}

	if err = os.MkdirAll(filepath.Base(path), 0755); err != nil {
		return pkgerrors.WithStack(err)
	}
	file, err := os.Create(path)
	if err != nil {
		return pkgerrors.WithStack(err)
	}

	return pkgerrors.WithStack(t.Format(file))
}
