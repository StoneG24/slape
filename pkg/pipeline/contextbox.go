package pipeline

import (
	"fmt"

	"github.com/StoneG24/slape/internal/vars"
	"github.com/openai/openai-go"
)

// ContextBox is a struct that contains a
// group of strings that contains context on a given problem.
// This is coupled with the system prompt chosen is what makes the models understand
// the gven situation more.

// This information should be kept within a pipeline for privacy and safety reasons.
type ContextBox struct {
	SystemPrompt          string
	Thoughts              string
	Prompt                string
	ConversationHistory   *[]string
	FutureQuestions       string
	InternetSearchResults *[]string
	ToolResults           *[]string
	//VectorStore           vectorstore.VectorStore{}
}

// PromptBuilder takes the ContextBox and builds the system prompt
func (c *ContextBox) PromptBuilder(previousAnswer string) error {

	// since we are operating on a parameter its
	// safer to create a local copy
	prevAns := previousAnswer

	// TODO(v,t) Go and gather the additional context from

	// TODO(v) vector store
	context := ""

	// minirag
	rag := ""

	// information generated as prelinary thoughts
	var additionalContex string
	if len(context) != 0 && len(rag) != 0 {
		additionalContex = context + rag
	} else {
		additionalContex = "None"
	}

	if len(previousAnswer) == 0 {
		prevAns = "None"
	}

	fmt.Println(c.Thoughts, additionalContex, prevAns)
	c.SystemPrompt = fmt.Sprintf(c.SystemPrompt, c.Thoughts, additionalContex, prevAns)

	// TODO(v) do something different for debate where we have question/idea and ask the hats after.
	return nil
}

// getThought is used to generate initial thoughts about a given question.
// This is supposed to create some guardrails for thought.
// This will not be good for slms but llms that are centered around reasoning
func (c *ContextBox) getThoughts() (string, error) {

	prompt := `You are an intellegent Small Language Model.
    You answer problems in a simple manner. 
    First think through this problem and return your thoughts.
    `
	param := openai.ChatCompletionNewParams{
		Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(prompt),
			openai.UserMessage(c.Prompt),
			//openai.UserMessage(s.FutureQuestions),
		}),
		Seed: openai.Int(0),
		//Model:       openai.String(pipeline.Model),
		Temperature: openai.Float(vars.ModelTemperature),
		//MaxTokens:   openai.Int(maxtokens),
	}

	result, err := GenerateCompletion(param, "", *vars.OpenaiClient)
	if err != nil {
		return "None", err
	}

	return result, nil
}
