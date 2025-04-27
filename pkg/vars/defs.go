package vars

import (
	"github.com/StoneG24/slape/pkg/prompt"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

const (
	CpuImage = "ghcr.io/ggml-org/llama.cpp:server"

	CudagpuImage = "ghcr.io/ggml-org/llama.cpp:server-cuda"
	RocmgpuImage = "ghcr.io/ggml-org/llama.cpp:server-rocm"

	Logfilename   = "logs.txt"
	Trunkfilename = "trunk.txt"

	// change to false to not run frontend
	Frontend = true

	// This should be used to match the context length with the max generation length.
	ContextLength      = 16348
	MaxGenTokens       = 16348
	MaxGenTokensSimple = 1024
	MaxGenTokensCoT    = 4096
	ModelTemperature   = 0.1

	// Timeout for generation (mins)
	GenerationTimeout = 10
)

var (
	OpenaiClient = openai.NewClient(
		option.WithBaseURL("http://localhost:8000/v1"),
	)

	GenerationClient = openai.NewClient(
		option.WithBaseURL("http://localhost:8081/v1"),
	)

	EmbeddingClient = openai.NewClient(
		option.WithBaseURL("http://localhost:8082/v1"),
	)

	ThinkingPrompt = prompt.ThinkingPrompt
	SimplePrompt   = prompt.SimplePrompt
	// todo sec prompts
	CotPrompt          = prompt.SecCoTPrompt
	TotPrompt          = prompt.SecToTPrompt
	GotPrompt          = prompt.SecGoTPrompt
	MoePrompt          = prompt.SecMoEPrompt
	ThinkingHatsPrompt = prompt.SecSixThinkingHats
	GoePrompt          = prompt.GoEPrompt
)
