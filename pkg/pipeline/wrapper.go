package pipeline

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strconv"

	"github.com/StoneG24/slape/pkg/vars"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/openai/openai-go"
)

// PullImage uses the docker api to pull an image down.
// The function also checks for the image locally before pulling.
func PullImage(apiClient *client.Client, ctx context.Context, containerImage string) (io.ReadCloser, error) {
	reader, err := apiClient.ImagePull(ctx, containerImage, image.PullOptions{All: false, RegistryAuth: ""})

	return reader, err
}

func CreateContainer(apiClient *client.Client, portNum string, name string, ctx context.Context, modelName string, containerImage string, gpuTrue bool) (container.CreateResponse, error) {

	portSet := nat.PortSet{
		nat.Port("8000/tcp"): struct{}{}, // map 11434 TCP port
	}

	portBindings := nat.PortMap{
		nat.Port("8000/tcp"): []nat.PortBinding{
			{
				HostIP:   "0.0.0.0",
				HostPort: portNum,
			},
		},
	}

	var mountString string

	if runtime.GOOS == "windows" {
		ex, err := os.Executable()
		if err != nil {
			slog.Error("idk something else")
		}

		currentPath := filepath.Dir(ex)

		slog.Debug(currentPath)

		mountString = currentPath + "\\models"
	}

	if runtime.GOOS == "linux" {
		mountString = os.Getenv("PWD") + "/models"
	}

	// TODO(v) add --jinja for function calling using the OpenAI API setup
	var cmds []string
	if gpuTrue {
		cmds = []string{"-m", "/models/" + modelName, "--port", "8000", "--host", "0.0.0.0", "-ngl", "-1", "-fa", "--no-webui", "-c", strconv.Itoa(vars.ContextLength), "-cb"}
	} else {
		cmds = []string{"-m", "/models/" + modelName, "--port", "8000", "--host", "0.0.0.0", "-fa", "--mlock", "--no-webui", "-c", strconv.Itoa(vars.ContextLength), "-cb"}
	}

	var hostconfig container.HostConfig

	// TODO(v) expand past nvidia systems.
	// ROCm will present interesting challenges. Its simpler but more setups in the config.
	switch gpuTrue {
	// BUG(v) This falls to the same problem as the rest of the gpu issues.
	case true:
		hostconfig = container.HostConfig{
			Runtime:      "nvidia",
			PortBindings: portBindings,
			Mounts: []mount.Mount{{
				Type:   mount.TypeBind,
				Source: mountString,
				Target: "/models",
			}},
		}
	case false:
		hostconfig = container.HostConfig{
			PortBindings: portBindings,
			Mounts: []mount.Mount{{
				Type:   mount.TypeBind,
				Source: mountString,
				Target: "/models",
			}},
		}
	}

	// create container
	createResponse, err := apiClient.ContainerCreate(ctx, &container.Config{
		ExposedPorts: portSet,
		Image:        containerImage,
		Cmd:          cmds,
	}, &hostconfig, nil, nil, name)

	return createResponse, err
}

// This is very simple for right now but when we add structured outputs it will
// get very complicated.
//
// prompt comes from a user and is the question being asked.
// systemprompt is the systemprompt chosen based on the prompting style requested.
func GenerateCompletion(ctx context.Context, param openai.ChatCompletionNewParams, followupQuestion string, openaiClient openai.Client) (string, error) {

	var result string

	stream := openaiClient.Chat.Completions.NewStreaming(ctx, param)

	// optionally, an accumulator helper can be used
	acc := openai.ChatCompletionAccumulator{}

	for stream.Next() {
		chunk := stream.Current()
		//w.Write([]byte(chunk.Choices[0].Delta.Content))
		acc.AddChunk(chunk)

		/*
			if content, ok := acc.JustFinishedContent(); ok {
				println("Content stream finished:", content)
			}
		*/

		// if using tool calls
		//if tool, ok := acc.JustFinishedToolCall(); ok {
		//	println("Tool call stream finished:", tool.Index, tool.Name, tool.Arguments)
		//}

		if refusal, ok := acc.JustFinishedRefusal(); ok {
			println("Refusal stream finished:", refusal)
		}

		// it's best to use chunks after handling JustFinished events
		if len(chunk.Choices) > 0 {
			print(chunk.Choices[0].Delta.Content)
		}
	}
	println("\n")

	if err := stream.Err(); err != nil {
		return "", err
	}

	// After the stream is finished, acc can be used like a ChatCompletion
	result = acc.Choices[0].Message.Content

	// Adding this for later
	//param.Messages.Value = append(param.Messages.Value, acc.Choices[0].Message)
	//param.Messages.Value = append(param.Messages.Value, openai.UserMessage(followupQuestion))

	return result, nil
}

// GenerateEmbedding is used as a helper function for generating embeddings.
func GenerateEmbedding(ctx context.Context, param openai.EmbeddingNewParams, client openai.Client) (openai.CreateEmbeddingResponse, error) {

	embeddings, err := client.Embeddings.New(ctx, param)
	if err != nil {
		return openai.CreateEmbeddingResponse{}, err
	}

	return *embeddings, nil
}
