package pipeline

// Should be created with openai spec.
// This seems to be the easiest way to make tools generic to the backend for models
// This also means its defined in code and less in raw json.
type Tool struct {
	// Descriptions are normally needed to
	// explain what the tool is how its used.
	Description string `json:"description"`
}

// Tools is used to define a list of tools availible to a pipeline
type Tools []string

func JsonifyTools() {}
