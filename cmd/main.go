/*
Package SLaPE is a binary that orchestrates containers using docker on the local computer.

Usage:

	./slape

Containerized models are spawned as needed adhering to a pipeline system.
*/

package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"time"

	"github.com/StoneG24/slape/pkg/api"
	"github.com/StoneG24/slape/pkg/logging"
	"github.com/StoneG24/slape/pkg/pipeline"
	"github.com/StoneG24/slape/pkg/vars"
	"github.com/docker/docker/client"
)

var (
	isGPU = pipeline.IsGPU()
	image = pipeline.PickImage()

	s = pipeline.SimplePipeline{
		// updates after created
		Models:         []string{},
		ContextBox:     pipeline.ContextBox{},
		Tools:          pipeline.Tools{},
		DockerClient:   nil,
		ContainerImage: image,
		GPU:            isGPU,
	}

	c = pipeline.ChainofModels{
		// updates after created
		Models:         []string{},
		ContextBox:     pipeline.ContextBox{},
		Tools:          pipeline.Tools{},
		DockerClient:   nil,
		ContainerImage: image,
		GPU:            isGPU,
	}

	d = pipeline.DebateofModels{
		// updates after created
		Models:         []string{},
		ContextBox:     pipeline.ContextBox{},
		Tools:          pipeline.Tools{},
		DockerClient:   nil,
		ContainerImage: image,
		GPU:            isGPU,
	}

	e = pipeline.EmbeddingPipeline{
		// updates after created
		DockerClient:   nil,
		ContainerImage: image,
		GPU:            isGPU,
	}
)

func main() {

	apiclient := createClient()

	s.DockerClient = apiclient
	c.DockerClient = apiclient
	d.DockerClient = apiclient
	e.DockerClient = apiclient

	logging.CreateLogFile()
	defer logging.CloseLogging()

	// Default Mux for our server.
	// For auth in the future we will want to setup a different set.
	mux := http.NewServeMux()

	mux.HandleFunc("POST /simple/generate", s.SimplePipelineGenerateRequest)
	mux.HandleFunc("POST /simple/setup", s.SimplePipelineSetupRequest)
	mux.HandleFunc("GET /simple/shutdown", s.Shutdown)
	mux.HandleFunc("POST /cot/generate", c.ChainPipelineGenerateRequest)
	mux.HandleFunc("POST /cot/setup", c.ChainPipelineSetupRequest)
	mux.HandleFunc("GET /cot/shutdown", c.Shutdown)
	mux.HandleFunc("POST /deb/setup", d.DebatePipelineSetupRequest)
	mux.HandleFunc("POST /deb/generate", d.DebatePipelineGenerateRequest)
	mux.HandleFunc("GET /deb/shutdown", d.Shutdown)
	mux.HandleFunc("GET  /emb/setup", e.EmbeddingPipelineSetupRequest)
	mux.HandleFunc("POST /emb/generate", e.EmbeddingPipelineGenerateRequest)
	mux.HandleFunc("GET /emb/shutdown", e.Shutdown)
	//mux.HandleFunc("/moe", simplerequest)
	//mux.HandleFunc("/up", upDog)
	mux.HandleFunc("GET /getmodels", api.GetModels)

	// This is against my religion
	wrappingMux := NewCoors(mux)

	// Create a new HTTP server.
	srv := &http.Server{
		Addr:    ":8080",
		Handler: wrappingMux,
	}

	// Start the server in a goroutine.
	log.Println("[+] Server started on :8080")
	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe(): %s", err)
		}
	}()

	// If we don't run embedding pipeline on startup,
	// we can remove this as well.
	log.Println("[+] Checking for models folder...")
	if _, err := os.Stat("./models"); errors.Is(err, os.ErrNotExist) {
		log.Println("[+] Creating models folder...")
		os.Mkdir("models", 1644)
	}

	// If needed, download the snowflake embedding model
	// If this is changed to gated model then the code would need to change to accept a token.
	// Reading from an evironment variable would be the safest option.
	if _, err := os.Stat("./models/snowflake-arctic-embed-l-v2.0-q4_k_m.gguf"); errors.Is(err, os.ErrNotExist) {
		log.Println("[+] Downloading Embedding Model...")
		err := downloadHuggingFaceModel(
			"Casual-Autopsy/snowflake-arctic-embed-l-v2.0-gguf",
			"snowflake-arctic-embed-l-v2.0-q4_k_m.gguf",
		)
		if err != nil {
			log.Fatalln("[-] Error Downloading Embedding Model", err)
		}
		log.Println("[+] Finished Downloading Embedding Model")
	}

	// starting up the embedding pipeline
	url := "http://localhost:8080/emb/setup"
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("[-] Error while trying to startup the Embedding Pipeline")
	}
	resp.Body.Close()

	// starting up the frontend on port 3000
	if vars.Frontend {
		log.Println("[+] Starting Frontend...")
		cmd := exec.Command("deno", "run", "dev")
		cmd.Dir = "./SLaMO_Frontend"
		go cmd.Run()
		log.Println("[+] Frontend has been started on port 3000")
	}

	// Create a channel to listen for interrupt signals.
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	// Block until a signal is received.
	<-sigChan

	// TODO(v) need to shutdown frontend process for windows
	// with every startup we spin up a server without tearing it down
	// windows would require a admin priv to remove the server by port like this
	log.Println("[+] Shuting Down Frontend...")
	if runtime.GOOS == "linux" {
		cmd := exec.Command("fuser", "-k", "3000/tcp")
		cmd.Run()
	}
	if runtime.GOOS == "windows" {
		//taskkill /f /pid $(netstat -ano | findstr ":3000")

		cmd := exec.Command("taskkill", "/f", "/pid", "$(netstat -ano | findstr ':3000')")
		cmd.Run()
	}
	log.Println("[+] Frontend has been stopped")

	err = shutdownPipelines()
	if err != nil {
		log.Println("ErrorShuttingDownPipelines:", err)
	}

    // clean up docker clients and free up sockets
	s.DockerClient.Close()
	c.DockerClient.Close()
	d.DockerClient.Close()
	e.DockerClient.Close()

	// Create a context with a timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Shutdown the server.
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Shutdown(): %s", err)
	}

	// Close the pipeline to stop adding new pipelines
	// close(keystone)

	log.Println("[+] Server gracefully stopped")
}

type Coors struct {
	handler http.Handler
}

// ServeHTTP handles the request by passing it to the real
// handler and logging the request details
func (c *Coors) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if origin := req.Header.Get("Origin"); origin != "" {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers",
			"Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	}
	// Stop here if its Preflighted OPTIONS request
	if req.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	c.handler.ServeHTTP(w, req)
}

// NewCoors constructs a new Logger middleware handler
func NewCoors(handlerToWrap http.Handler) *Coors {
	return &Coors{handlerToWrap}
}

// shutdownPipelines is used to shutdown pipelines with
// a remote request since the shutdown functions are now http.HandleFunc
func shutdownPipelines() error {

	url := "http://localhost:8080/%s/shutdown"
	pipelines := []string{"simple", "cot", "deb", "emb"}

	for _, pipeline := range pipelines {
		requrl := fmt.Sprintf(url, pipeline)
		resp, err := http.Get(requrl)
		if err != nil {
			return err
		}
		resp.Body.Close()
	}

	return nil
}

// DownloadHuggingFaceModel downloads a given model provided a repo and filename are given.
// This only really works for our usecase since we are using a gguf model.
// Note This functionality is already in llamacpp-server
func downloadHuggingFaceModel(repo string, filename string) error {
	url := "https://huggingface.co/" + repo + "/resolve/main/" + filename

	// Send HTTP GET request
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check if status is OK (200)
	if resp.StatusCode != http.StatusOK {
		return err
	}

	// Create the file
	out, err := os.Create("./models/" + filename)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write response body to file
	_, err = io.Copy(out, resp.Body)
	return err
}

func createClient() *client.Client {
	apiClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Println("Error creating the docker client: ", err)
		return nil
	}

	return apiClient
}
