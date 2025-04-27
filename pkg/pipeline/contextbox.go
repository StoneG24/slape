package pipeline

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/StoneG24/slape/pkg/api"
	"github.com/StoneG24/slape/pkg/internetsearch"
	"github.com/StoneG24/slape/pkg/vars"
	"github.com/openai/openai-go"
)

// ContextBox is a struct that contains a
// group of strings that contains context on a given problem.
// This is coupled with the system prompt chosen is what makes the models understand
// the gven situation more.
// This information should be kept within a pipeline for privacy and safety reasons.
type ContextBox struct {
	// Simple prompt components
	SystemPrompt string
	Thoughts     string
	Prompt       string
	// Currently not in use
	ConversationHistory *[]string
	FutureQuestions     string

	// These will come from the internet search package.
	InternetSearchResults string
	//VectorStore           *hnsw.Graph[string]

	// These will come from tool calls
	ToolResults *[]string
}

// PromptBuilder takes the ContextBox and builds the system prompt
func (c *ContextBox) promptBuilder(previousAnswer string) error {

	// since we are operating on a parameter its
	// safer to create a local copy
	prevAns := previousAnswer
	if len(previousAnswer) == 0 {
		prevAns = "None"
	}

	// information generated as prelinary thoughts
	// TODO(v) move to generation functions like thoughts
	var additionalContex string
	if len(c.InternetSearchResults) != 0 {
		additionalContex += c.InternetSearchResults
	} else {
		additionalContex = "None"
	}

	log.Println(c.Thoughts, additionalContex, prevAns)
	c.SystemPrompt = fmt.Sprintf(c.SystemPrompt, c.Thoughts, additionalContex, prevAns)

	// TODO(v) do something different for debate where we have question/idea and ask the hats after.
	return nil
}

// getThought is used to generate initial thoughts about a given question.
// This is supposed to create some guardrails for thought.
// This will not be good for slms but llms that are centered around reasoning
func (c *ContextBox) getThoughts(ctx context.Context) {

    fmt.Println(c.InternetSearchResults)
	prompt := vars.ThinkingPrompt + "\n**Internet Search Results:**\n" + c.InternetSearchResults

	param := openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(prompt),
			openai.UserMessage(c.Prompt),
			//openai.UserMessage(s.FutureQuestions),
		},
		Seed: openai.Int(0),
		//Model:       openai.String(pipeline.Model),
		Temperature: openai.Float(0.4),
		MaxTokens:   openai.Int(vars.MaxGenTokens),
	}

    fmt.Println(param.Messages)

	for {
		// sleep and give server guy a break
		time.Sleep(time.Duration(1 * time.Second))

		// Single model, single port, assuming one pipeline is running at a time
		if api.UpDog("8000") {
			break
		}
	}

	result, err := GenerateCompletion(ctx, param, "", vars.OpenaiClient)
	log.Println(result)
	if err != nil {
		c.Thoughts = "None"
	}

	log.Println("Debug Thinking result", result)

	c.Thoughts = result
}

// getInternetSearch is used to generate initial context about a given question.
func (c *ContextBox) getInternetSearch(ctx context.Context) error {

	// this gets converted to f32 with iterative process
	embCh := make(chan []float64, 1)
	searchCh := make(chan internetsearch.VectorList, 1)
	// defer close(embCh)
	//defer close(searchCh)

	// Generate embedding of prompt
	go func(context.Context, chan []float64) {
		embedparam := openai.EmbeddingNewParams{
			Input:      openai.EmbeddingNewParamsInputUnion{OfArrayOfStrings: []string{c.Prompt}},
			Model:      embedmodel,
			Dimensions: openai.Int(1024),
		}

		for {
			// sleep and give server guy a break
			time.Sleep(time.Duration(2 * time.Second))

			// Single model, single port, assuming one pipeline is running at a time
			if api.UpDog("8082") {
				break
			}
		}

		result, err := GenerateEmbedding(ctx, embedparam, vars.EmbeddingClient)
		log.Println(result)
		if err != nil {
			return
		}
		embCh <- result.Data[0].Embedding
		//close(embCh)
	}(ctx, embCh)

	// Generate the query and run internetsearch
	go func(context.Context, chan internetsearch.VectorList) {
		// take care of upDog on our own
		for {
			// sleep and give server guy a break
			time.Sleep(time.Duration(1 * time.Second))

			// Single model, single port, assuming one pipeline is running at a time
			if api.UpDog("8000") {
				break
			}
		}

		/*
		   queryPrompt := `
		   Act as a internet search guru who knows how to search up anything on duckduckgo.com.
		   Generate a query, using the provided question, for searching the internet.
		   Only return the question.
		   `
		*/
		// take the result and run the internetsearch
		vecs := internetsearch.InternetSearch(ctx, c.Prompt)

		searchCh <- vecs
		//close(searchCh)
	}(ctx, searchCh)

	// Combine the two and search the graph for relative neighbors
	var embedding []float64
	var vecs internetsearch.VectorList
	for i := 0; i < 2; i++ {
		select {
		case vector, ok := <-embCh:
			if ok {
				embedding = vector
				log.Println("Embedding Retrieved Properly")
			}
		case v, ok := <-searchCh:
			if ok {
				vecs = v
				log.Println("Neighbors Retrieved Properly")
			}
		}
	}

	neighbors := internetsearch.KnnSearch(
		// nearest neighbors
		vecs.Points,
		// embedding vector
		embedding,
		// change this to get less results back from the vector store
		5,
	)

	log.Println("Internet Search result [nearest neighbors]", neighbors)

	for _, neighbor := range neighbors {
		c.InternetSearchResults += vecs.Elements[neighbor.Point.ID]
		c.InternetSearchResults += "\n"
	}

	log.Println("Internet Search result ", c.InternetSearchResults)

	return nil
}
