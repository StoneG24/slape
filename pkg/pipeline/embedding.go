package pipeline

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/StoneG24/slape/internal/vars"
	"github.com/StoneG24/slape/pkg/api"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/fatih/color"
	"github.com/openai/openai-go"
)

const (
	embedmodel = "snowflake-arctic-embed-l-v2.0-q4_k_m.gguf"
	genmodel   = "Phi-3.5-mini-instruct-Q4_K_M.gguf"
)

type (
	// This pipeline is meant to be used for indexing a RAG database.
	// We are using MiniRAG for a size complexity balance.
	EmbeddingPipeline struct {
		DockerClient   *client.Client
		ContainerImage string
		GPU            bool

		// for internal use
		gencontainer container.CreateResponse
		embcontainer container.CreateResponse
	}

	embeddingResponse struct {
		Response openai.CreateEmbeddingResponse
	}
)

// SimplePipelineSetupRequest, handlerfunc expects POST method and returns no content
func (e *EmbeddingPipeline) EmbeddingPipelineSetupRequest(w http.ResponseWriter, req *http.Request) {
	apiClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		color.Red("%s", err)
		return
	}
	go api.Cors(w, req)

	// setup values needed for pipeline
	e.DockerClient = apiClient

	go e.Setup(context.Background())

	w.WriteHeader(http.StatusOK)
}

// simplerequest is used to handle simple requests as needed.
func (e *EmbeddingPipeline) EmbeddingPipelineGenerateRequest(w http.ResponseWriter, req *http.Request) {
	ctx := context.Background()
	go api.Cors(w, req)

	var simplePayload simpleRequest

	err := json.NewDecoder(req.Body).Decode(&simplePayload)
	if err != nil {
		color.Red("%s", err)
		http.Error(w, "Error unexpected request format", http.StatusUnprocessableEntity)
		return
	}

	// generate a response
	// TODO rewrite for embedding and rag
	result, err := e.Generate(simplePayload.Prompt, vars.EmbeddingClient)
	if err != nil {
		color.Red("%s", err)
		http.Error(w, "Error getting generation from model", http.StatusOK)
		go e.Shutdown(ctx)

		return
	}

	go e.Shutdown(ctx)

	// for debugging streaming
	color.Green("%s", result)

	respPayload := embeddingResponse{
		Response: *result,
	}

	json, err := json.Marshal(respPayload)
	if err != nil {
		color.Red("%s", err)
		http.Error(w, "Error marshaling your response from model", http.StatusOK)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(json)
}

func (e *EmbeddingPipeline) Setup(ctx context.Context) error {

	reader, err := PullImage(e.DockerClient, ctx, e.ContainerImage)
	if err != nil {
		color.Red("%s", err)
		return err
	}
	color.Green("Pulling Image...")
	// prints out the status of the download
	// worth while for big images
	io.Copy(os.Stdout, reader)
	defer reader.Close()

	gencreateResponse, err := CreateContainer(
		e.DockerClient,
		"8081",
		"",
		ctx,
		genmodel,
		e.ContainerImage,
		e.GPU,
	)

	embedcreateResponse, err := CreateContainer(
		e.DockerClient,
		"8082",
		"",
		ctx,
		embedmodel,
		e.ContainerImage,
		e.GPU,
	)

	if err != nil {
		color.Yellow("%s", gencreateResponse.Warnings)
		color.Yellow("%s", embedcreateResponse.Warnings)
		color.Red("%s", err)
		return err
	}

	// start container
	err = (e.DockerClient).ContainerStart(context.Background(), gencreateResponse.ID, container.StartOptions{})
	if err != nil {
		color.Red("%s", err)
		return err
	}

	// start container
	err = (e.DockerClient).ContainerStart(context.Background(), embedcreateResponse.ID, container.StartOptions{})
	if err != nil {
		color.Red("%s", err)
		return err
	}

	// For debugging
	color.Green("%s", gencreateResponse.ID)
	color.Green("%s", embedcreateResponse.ID)
	e.embcontainer = embedcreateResponse
	e.gencontainer = gencreateResponse

	return nil
}

func (e *EmbeddingPipeline) Generate(payload string, openaiClient *openai.Client) (*openai.CreateEmbeddingResponse, error) {
	// take care of upDog on our own
	for {
		// sleep and give server guy a break
		time.Sleep(time.Duration(5 * time.Second))

		// Single model, single port, assuming one pipeline is running at a time
		if api.UpDog("8081") && api.UpDog("8082") {
			break
		}
	}

	param := openai.EmbeddingNewParams{
		Input:      openai.F(openai.EmbeddingNewParamsInputUnion(openai.EmbeddingNewParamsInputArrayOfStrings{payload})),
		Model:      openai.String(embedmodel),
		Dimensions: openai.Int(1024),
	}

	// should return a type of openai.Embedding
	result, err := GenerateEmbedding(param, *openaiClient)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (e *EmbeddingPipeline) Shutdown(ctx context.Context) error {
	err := (e.DockerClient).ContainerStop(ctx, e.gencontainer.ID, container.StopOptions{})
	if err != nil {
		color.Red("%s", err)
		return nil
	}

	err = (e.DockerClient).ContainerStop(ctx, e.embcontainer.ID, container.StopOptions{})
	if err != nil {
		color.Red("%s", err)
		return nil
	}

	err = (e.DockerClient).ContainerRemove(ctx, e.gencontainer.ID, container.RemoveOptions{})
	if err != nil {
		color.Red("%s", err)
		return nil
	}

	err = (e.DockerClient).ContainerRemove(ctx, e.embcontainer.ID, container.RemoveOptions{})
	if err != nil {
		color.Red("%s", err)
		return nil
	}

	color.Green("Shutting Down...")

	return nil
}
