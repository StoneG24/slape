package pipeline

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/StoneG24/slape/pkg/api"
	"github.com/StoneG24/slape/pkg/vars"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
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
		// 0 is embedding model
		// 1 is generation model
		containers []container.CreateResponse
	}

	EmbeddingRequest struct {
		Prompt []string `json:"prompt"`
	}

	EmbeddingResponse struct {
		Response openai.CreateEmbeddingResponse
	}
)

// SimplePipelineSetupRequest, handlerfunc expects POST method and returns no content
func (e *EmbeddingPipeline) EmbeddingPipelineSetupRequest(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	apiClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Println("Error creating the docker client: ", err)
		return
	}

	// setup values needed for pipeline
	e.DockerClient = apiClient

	e.Setup(ctx)

	w.WriteHeader(http.StatusOK)
}

// simplerequest is used to handle simple requests as needed.
func (e *EmbeddingPipeline) EmbeddingPipelineGenerateRequest(w http.ResponseWriter, req *http.Request) {
	var payload EmbeddingRequest

	ctx := req.Context()

	err := json.NewDecoder(req.Body).Decode(&payload)
	if err != nil {
		log.Println("Error Request Format: ", err)
		http.Error(w, "Error unexpected request format", http.StatusUnprocessableEntity)
		return
	}

	// generate a response
	// TODO rewrite for embedding and rag
	result, err := e.Generate(ctx, payload.Prompt, &vars.EmbeddingClient)
	if err != nil {
		log.Println("Error getting generation from model", err)
		http.Error(w, "Error getting generation from model", http.StatusInternalServerError)

		return
	}

	respPayload := EmbeddingResponse{
		Response: *result,
	}

	json, err := json.Marshal(respPayload)
	if err != nil {
		log.Println("Error marshaling response from model", err)
		http.Error(w, "Error marshaling your response from model", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(json)
}

func (e *EmbeddingPipeline) Setup(ctx context.Context) error {

	log.Println("PullingImage: ", e.ContainerImage)

	reader, err := PullImage(e.DockerClient, ctx, e.ContainerImage)
	if err != nil {
		log.Println("Error Pulling Container Image: ", err)
		return err
	}
	// prints out the status of the download
	// worth while for big images
	io.Copy(os.Stdout, reader)
	defer reader.Close()

	/*
		gencreateResponse, err := CreateContainer(
			e.DockerClient,
			"8081",
			"",
			ctx,
			genmodel,
			e.ContainerImage,
			e.GPU,
		)
	*/

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
		log.Println("Create Container Warning: ", embedcreateResponse.Warnings)
		log.Println("Error Creating Container: ", err)
		return err
	}

	// start container
	/*
		err = (e.DockerClient).ContainerStart(ctx, gencreateResponse.ID, container.StartOptions{})
		if err != nil {
			slog.Error("Error", "Errostring", err)
			return err
		}
	*/

	// start container
	err = (e.DockerClient).ContainerStart(ctx, embedcreateResponse.ID, container.StartOptions{})
	if err != nil {
		log.Println("Error Starting Container: ", err)
		return err
	}

	//slog.Info("%s", gencreateResponse.ID)
	log.Println("Starting Container: ", embedcreateResponse.ID)

	e.containers = append(e.containers, embedcreateResponse)
	//e.containers = append(e.containers, gencreateResponse)

	return nil
}

func (e *EmbeddingPipeline) Generate(ctx context.Context, payload []string, openaiClient *openai.Client) (*openai.CreateEmbeddingResponse, error) {
	// take care of upDog on our own
	for {
		// sleep and give server guy a break
		time.Sleep(time.Duration(2 * time.Second))

		// Single model, single port, assuming one pipeline is running at a time
		//if api.UpDog("8081") && api.UpDog("8082") {
		if api.UpDog("8082") {

			break
		}
	}

	param := openai.EmbeddingNewParams{
		Input:      openai.EmbeddingNewParamsInputUnion{OfArrayOfStrings: payload},
		Model:      embedmodel,
		Dimensions: openai.Int(1024),
	}

	// should return a type of openai.Embedding
	result, err := GenerateEmbedding(ctx, param, *openaiClient)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (e *EmbeddingPipeline) Shutdown(w http.ResponseWriter, req *http.Request) {

	childctx, cancel := context.WithDeadline(req.Context(), time.Now().Add(30*time.Second))
	defer cancel()

	// turn off the containers if they aren't already off
	for _, model := range e.containers {
		(e.DockerClient).ContainerStop(childctx, model.ID, container.StopOptions{})
	}

	// remove the containers, seperate incase it's already stopped
	for _, model := range e.containers {
		(e.DockerClient).ContainerRemove(childctx, model.ID, container.RemoveOptions{})
	}

	log.Println("Shutting Down...")
}
