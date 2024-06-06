go install -v github.com/symflower/eval-dev-quality/...

eval-dev-quality.exe evaluate `
        --runs 5 `
        --runs-sequential `
        --result-path .\light-full-gpu `
        --repository java\light `
        --repository golang\light `
        --model ollama/codeqwen:7b-code `
        --model ollama/codeqwen:7b-chat `
        --model ollama/aya:8b `
        --model ollama/starcoder2:3b `
        --model ollama/starcoder2:7b `
        --model ollama/yi:6b-v1.5 `
        --model ollama/yi:9b-v1.5 `
        --model ollama/granite-code:3b `
        --model ollama/granite-code:8b