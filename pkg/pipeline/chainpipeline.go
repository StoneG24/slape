package pipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/StoneG24/slape/pkg/api"
	"github.com/StoneG24/slape/pkg/prompt"
	"github.com/StoneG24/slape/pkg/vars"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

type (
	// ChainofModels is the next step above smallest pipeline.
	// This pipeline contains a ContextBox and the models in squential order.
	// ChainofModels forces the models to talk in sequential order
	// like the name suggests.
	ChainofModels struct {
		Models         []string
		ContainerImage string
		Thinking       bool
		InternetSearch bool
		GPU            bool
		DockerClient   *client.Client

		// embedded structs
		ContextBox
		Tools

		// for internal use to store the models in
		containers []container.CreateResponse
	}

	chainRequest struct {
		// Prompt is the string that
		// will be appended to the prompt
		// string chosen.
		Prompt string `json:"prompt"`

		// Options are strings matching
		// the names of prompt types
		Mode string `json:"mode"`

		// Should we have a thinking step involved
		Thinking string `json:"thinking"`

		// Should we have a thinking step involved
		InternetSearch string `json:"search"`
	}

	chainSetupPayload struct {
		Models []string `json:"models"`
	}

	chainResponse struct {
		Answer string `json:"answer"`
	}
)

// ChainPipelineSetupRequest, expects POST method and returns nothing. Runs the startup
// process for a chain pipeline.
func (c *ChainofModels) ChainPipelineSetupRequest(w http.ResponseWriter, req *http.Request) {
	var setupPayload chainSetupPayload

	ctx, cancel := context.WithDeadline(req.Context(), time.Now().Add(30*time.Second))
	defer cancel()

	err := json.NewDecoder(req.Body).Decode(&setupPayload)
	if err != nil {
		log.Println("Error Request Format: ", err)
		http.Error(w, "Error unexpected request format", http.StatusUnprocessableEntity)
		return
	}

	c.Models = setupPayload.Models

	c.Setup(ctx)

	w.WriteHeader(http.StatusOK)
}

// ChainPipelineRequest is used to handle requests for chain of models pipelines.
// The json expected is
func (c *ChainofModels) ChainPipelineGenerateRequest(w http.ResponseWriter, req *http.Request) {
	var payload chainRequest

	// use this to scope the context to the request
	ctx, cancel := context.WithDeadline(req.Context(), time.Now().Add(vars.GenerationTimeout*time.Minute))
	defer cancel()

	err := json.NewDecoder(req.Body).Decode(&payload)
	if err != nil {
		log.Println("Error Request Format", err)
		http.Error(w, "Error unexpected request format", http.StatusUnprocessableEntity)
		return
	}

	promptChoice, maxtokens := processPrompt(payload.Mode)

	c.ContextBox.SystemPrompt = promptChoice
	c.ContextBox.Prompt = payload.Prompt
	c.Thinking, err = strconv.ParseBool(payload.Thinking)
	if err != nil {
		log.Println("Error Parsing thinking value:", err)
		http.Error(w, "Error parsing thinking value. Expecting sound boolean definitions.", http.StatusBadRequest)
	}
	c.InternetSearch, err = strconv.ParseBool(payload.InternetSearch)
	if err != nil {
		log.Println("Error Parsing InternetSearch value:", err)
		http.Error(w, "Error parsing InternetSearch value. Expecting sound boolean definitions.", http.StatusBadRequest)
	}

	if c.InternetSearch {
		c.getInternetSearch(ctx)
	} else {
		c.InternetSearchResults = []string{}
	}
	if c.Thinking {
		c.getThoughts(ctx)
	} else {
		c.Thoughts = "None"
	}

	// wait on go routines then generate a response
	result, err := c.Generate(ctx, payload.Prompt, promptChoice, maxtokens)
	if err != nil {
		log.Println("Error getting generation from model", err)
		http.Error(w, "Error getting generation from model", http.StatusInternalServerError)
		return
	}

	// for debugging streaming
	log.Println(result)

	respPayload := chainResponse{
		Answer: result,
	}

	json, err := json.Marshal(respPayload)
	if err != nil {
		log.Println("Error marshaling response from model", err)
		http.Error(w, "Error marshaling your response from model", http.StatusInternalServerError)
		return
	}

	c.InternetSearchResults = []string{}
	c.Thoughts = ""

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(json)
}

func (c *ChainofModels) Setup(ctx context.Context) error {

	childctx, cancel := context.WithDeadline(ctx, time.Now().Add(30*time.Second))
	defer cancel()

	/*
		log.Println("PullingImage: ", c.ContainerImage)

		reader, err := PullImage(c.DockerClient, childctx, c.ContainerImage)
		if err != nil {
			log.Println("Error Pulling Docker Image for Containers", err)
			return err
		}
		// prints out the status of the download
		// worth while for big images
		io.Copy(os.Stdout, reader)
	*/

	for i, model := range c.Models {
		createResponse, err := CreateCPPContainer(
			c.DockerClient,
			"800"+strconv.Itoa(i),
			"",
			childctx,
			model,
			c.ContainerImage,
			c.GPU,
		)

		if err != nil {
			log.Println("Create Container Warning", createResponse.Warnings)
			log.Println("Error Creating Container", err)
			return err
		}

		log.Println("Container Created With ID", createResponse.ID)
		c.containers = append(c.containers, createResponse)
	}

	// start container
	err := (c.DockerClient).ContainerStart(childctx, c.containers[0].ID, container.StartOptions{})
	if err != nil {
		log.Println("Error Starting Container: ", err)
		return err
	}
	log.Println("Starting Container", c.containers[0].ID)

	return nil
}

// ChainofModels.Generate is the facilitator of model orchestration based on the chain of model pipeline.
// Since the pipeline is based on the Chan of Thought prompting technique, it follows this style, mimicing its behavior.
func (c *ChainofModels) Generate(ctx context.Context, uprompt string, systemprompt string, maxtokens int64) (string, error) {
	var result string

	for i, model := range c.containers {
		fmt.Println("ContainerIndex ", i)
		// start container
		err := (c.DockerClient).ContainerStart(ctx, model.ID, container.StartOptions{})
		if err != nil {
			log.Println("Error Starting Container", err)
			return "", err
		}
		log.Println("ContainerIndex ", i)

		for {
			// sleep and give server guy a break
			time.Sleep(time.Duration(1 * time.Second))

			if api.UpDog("800" + strconv.Itoa(i)) {
				break
			}
		}

		openaiClient := openai.NewClient(
			option.WithBaseURL("http://localhost:800" + strconv.Itoa(i) + "/v1"),
		)

		err = c.promptBuilder()
		if err != nil {
			return "", err
		}

		// Answer the initial question.
		// If it's the first model, there will not be any questions from the previous model.
		param := openai.ChatCompletionNewParams{
			Messages: []openai.ChatCompletionMessageParamUnion{
				openai.SystemMessage(c.SystemPrompt),
				openai.UserMessage(c.Prompt),
			},
			Seed:        openai.Int(0),
			Model:       c.Models[i],
			Temperature: openai.Float(vars.ModelTemperature),
			MaxTokens:   openai.Int(maxtokens),
		}

		// ans the question
		result, err = GenerateCompletion(ctx, param, "", openaiClient)
		if err != nil {
			log.Println("Error Generating Completion", err)
			return "", err
		}

		err = c.promptBuilder()
		if err != nil {
			return "", err
		}
		// after we have our values set we can clear out old ones to re-used
		c.FutureQuestions = "None"

		// if its not the last model summarize the response and generate more questions.
		if i != len(c.containers)-1 {
			// Summarize the answer generate.
			// This apparently makes it easier for the next models to digest the information.
			summarizePrompt := fmt.Sprintf(prompt.SummarizingPrompt, result)
			param = openai.ChatCompletionNewParams{
				Messages: []openai.ChatCompletionMessageParamUnion{
					openai.SystemMessage(prompt.SimplePrompt),
					openai.UserMessage(summarizePrompt),
					//openai.UserMessage(s.FutureQuestions),
				},
				Seed:        openai.Int(0),
				Model:       c.Models[i],
				Temperature: openai.Float(vars.ModelTemperature),
				MaxTokens:   openai.Int(maxtokens),
			}

			result, err = GenerateCompletion(ctx, param, "", openaiClient)
			if err != nil {
				log.Println("Error Generating Completion", err)
				return "", err
			}

			c.ConversationHistory = append(c.ConversationHistory, result)
			err = c.promptBuilder()
			if err != nil {
				return "", err
			}

			// Ask the model to generate questions for the model to answer.
			// Then store this answer in the contextbox for the next go around.
			qprompt := `Given this answer, %s, can you generate five questions to ask someone else, that pertain to the question?
            Your goal is to be concise while still maintaining the orginal message.`
			askFutureQuestions := fmt.Sprintf(qprompt, result)
			param = openai.ChatCompletionNewParams{
				Messages: []openai.ChatCompletionMessageParamUnion{
					openai.SystemMessage(c.SystemPrompt),
					openai.UserMessage(askFutureQuestions),
					//openai.UserMessage(s.FutureQuestions),
				},
				Seed:        openai.Int(0),
				Model:       c.Models[i],
				Temperature: openai.Float(vars.ModelTemperature),
				MaxTokens:   openai.Int(maxtokens),
			}

			result, err = GenerateCompletion(ctx, param, "", openaiClient)
			if err != nil {
				log.Println("Error Generating Completion", err)
				return "", err
			}

			c.FutureQuestions = result
			err = c.promptBuilder()
			if err != nil {
				return "", err
			}

			//log.Println("SystemPrompt", systemprompt, "Prompt", uprompt)
		}

		log.Println("Stopping Container, ContainerIndex", i)
		(c.DockerClient).ContainerStop(ctx, model.ID, container.StopOptions{})
	}

	c.ConversationHistory = []string{}

	// start container
	err := (c.DockerClient).ContainerStart(ctx, c.containers[0].ID, container.StartOptions{})
	if err != nil {
		log.Println("Error Starting Container: ", err)
		return result, err
	}
	log.Println("Starting Container For next Potential run", c.containers[0].ID)

	return result, nil
}

// ChainofModels.Shutdown handles the shutdown of the pipelines models.
func (c *ChainofModels) Shutdown(w http.ResponseWriter, req *http.Request) {

	childctx, cancel := context.WithDeadline(req.Context(), time.Now().Add(30*time.Second))
	defer cancel()

	// turn off the containers if they aren't already off
	for _, model := range c.containers {
		(c.DockerClient).ContainerStop(childctx, model.ID, container.StopOptions{})
	}

	// remove the containers, seperate incase it's already stopped
	for _, model := range c.containers {
		(c.DockerClient).ContainerRemove(childctx, model.ID, container.RemoveOptions{})
	}

	log.Println("Shutting Down...")
}
