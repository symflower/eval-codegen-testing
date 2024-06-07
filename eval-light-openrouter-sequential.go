package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	pkgerrors "github.com/pkg/errors"
	"github.com/zimmski/osutil"

	"github.com/symflower/eval-dev-quality/model"
	"github.com/symflower/eval-dev-quality/util"
)

var models = []string{
	"openrouter/01-ai/yi-34b",
	"openrouter/01-ai/yi-34b-chat",
	"openrouter/01-ai/yi-6b",
	"openrouter/allenai/olmo-7b-instruct",
	"openrouter/alpindale/goliath-120b",
	"openrouter/anthropic/claude-3-haiku",
	"openrouter/anthropic/claude-3-haiku:beta",
	"openrouter/anthropic/claude-3-opus",
	"openrouter/anthropic/claude-3-opus:beta",
	"openrouter/anthropic/claude-3-sonnet",
	"openrouter/anthropic/claude-3-sonnet:beta",
	"openrouter/anthropic/claude-instant-1.1",
	"openrouter/austism/chronos-hermes-13b",
	"openrouter/codellama/codellama-70b-instruct",
	"openrouter/cognitivecomputations/dolphin-mixtral-8x7b",
	"openrouter/cohere/command-r-plus",
	"openrouter/databricks/dbrx-instruct",
	"openrouter/deepseek/deepseek-chat",
	"openrouter/deepseek/deepseek-coder",
	"openrouter/fireworks/firellava-13b",
	"openrouter/google/gemini-flash-1.5",
	"openrouter/google/gemini-pro-1.5",
	"openrouter/google/gemma-7b-it",
	"openrouter/google/palm-2-chat-bison",
	"openrouter/google/palm-2-codechat-bison",
	"openrouter/huggingfaceh4/zephyr-7b-beta",
	"openrouter/intel/neural-chat-7b",
	"openrouter/jondurbin/airoboros-l2-70b",
	"openrouter/liuhaotian/llava-13b",
	"openrouter/liuhaotian/llava-yi-34b",
	"openrouter/meta-llama/codellama-34b-instruct",
	"openrouter/meta-llama/llama-3-70b",
	"openrouter/meta-llama/llama-3-8b",
	"openrouter/microsoft/phi-3-medium-128k-instruct",
	"openrouter/microsoft/phi-3-mini-128k-instruct",
	"openrouter/microsoft/wizardlm-2-7b",
	"openrouter/microsoft/wizardlm-2-8x22b",
	"openrouter/mistralai/mistral-7b-instruct-v0.3",
	"openrouter/mistralai/mistral-large",
	"openrouter/mistralai/mistral-medium",
	"openrouter/mistralai/mistral-small",
	"openrouter/mistralai/mistral-tiny",
	"openrouter/mistralai/mixtral-8x22b",
	"openrouter/mistralai/mixtral-8x7b",
	"openrouter/neversleep/llama-3-lumimaid-70b",
	"openrouter/neversleep/llama-3-lumimaid-8b",
	"openrouter/nousresearch/hermes-2-pro-llama-3-8b",
	"openrouter/nousresearch/nous-capybara-34b",
	"openrouter/nousresearch/nous-capybara-7b",
	"openrouter/nousresearch/nous-hermes-2-mistral-7b-dpo",
	"openrouter/nousresearch/nous-hermes-2-mixtral-8x7b-dpo",
	"openrouter/nousresearch/nous-hermes-2-mixtral-8x7b-sft",
	"openrouter/nousresearch/nous-hermes-yi-34b",
	"openrouter/open-orca/mistral-7b-openorca",
	"openrouter/openai/gpt-4",
	"openrouter/openai/gpt-4-turbo",
	"openrouter/openai/gpt-4o",
	"openrouter/perplexity/llama-3-sonar-large-32k-chat",
	"openrouter/perplexity/llama-3-sonar-small-32k-chat",
	"openrouter/phind/phind-codellama-34b",
	"openrouter/qwen/qwen-110b-chat",
	"openrouter/qwen/qwen-14b-chat",
	"openrouter/qwen/qwen-32b-chat",
	"openrouter/qwen/qwen-4b-chat",
	"openrouter/qwen/qwen-72b-chat",
	"openrouter/qwen/qwen-7b-chat",
	"openrouter/recursal/eagle-7b",
	"openrouter/recursal/rwkv-5-3b-ai-town",
	"openrouter/rwkv/rwkv-5-world-3b",
	"openrouter/snowflake/snowflake-arctic-instruct",
	"openrouter/teknium/openhermes-2-mistral-7b",
	"openrouter/teknium/openhermes-2.5-mistral-7b",
	"openrouter/togethercomputer/stripedhyena-hessian-7b",
	"openrouter/togethercomputer/stripedhyena-nous-7b",
	"openrouter/undi95/toppy-m-7b",
	"openrouter/xwin-lm/xwin-lm-70b",
}

func main() {
	progress := osutil.ProgressBar(os.Stdout, len(models), "Benchmarking")
	defer progress.Exit()

	for _, model := range models {
		progress.Describe("Benchmarking: " + model)

		err := benchmarkModel("evaluation-light-openrouter-sequential", "light", model, 5*time.Minute)
		if err != nil {
			fmt.Printf("Model %q misbehaved", model)
		}

		progress.Add(1)
	}
}

func benchmarkModel(resultPrefix string, repository string, id string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	output, err := util.CommandWithResult(ctx, log.New(io.Discard, "", 0), &util.Command{
		Command: []string{
			"eval-dev-quality",
			"--runs", "5",
			"--result-path", resultPrefix + model.CleanModelNameForFileSystem(id),
			"--repository", filepath.Join("java", repository),
			"--repository", filepath.Join("golang", repository),
			"--model", id,
		},
	})
	if err != nil {
		return pkgerrors.WithStack(pkgerrors.WithMessage(err, output))
	}

	return nil
}
