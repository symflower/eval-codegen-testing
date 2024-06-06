go install -v github.com/symflower/eval-dev-quality/...

eval-dev-quality.exe evaluate `
        --runs 5 `
        --runs-sequential `
        --result-path .\light-partial-gpu `
        --repository java\light `
        --repository golang\light `
        --model ollama/granite-code:20b `
        --model ollama/granite-code:34b `
        --model ollama/yi:34b-v1.5 `
        --model ollama/starcoder2:15b `
        --model ollama/codestral:22b `
        --model ollama/falcon2:11b `
        --model ollama/aya:35b