// Package api is used to create general functions needed for the api of slape.
// Some of which are defined in the pipelines and therefore are not included in this package.
package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
)

func UpDog(port string) bool {
	// swagger:operation GET /updog UpDog
	//
	// UpDog is a check to make sure llamacpp is ready for requests.
	//
	// Responses:
	//	200: StatusOk
	resp, err := http.Get("http://localhost:" + port + "/health")
	if err != nil {
		slog.Error("%s", err)
		return false
	}

	switch resp.StatusCode {
	case http.StatusOK:
		slog.Info("Model is ready")
		return true
	case http.StatusServiceUnavailable:
		slog.Warn("Model is not ready...")
		return false
	default:
		slog.Warn("Model is not ready...")
		return false
	}
}

// swagger:model ModelsResponse
type ModelsResponse struct {
	// List of models to be returned to the client
	// required: true
	Models []string `json:"models"`
}

// swagger:route GET /getmodels
//
// get models that the backend can see in the ./models folder
//
// Responses:
//
//	200: ModelsResponse
func GetModels(w http.ResponseWriter, req *http.Request) {
	files, err := os.ReadDir("./models")
	if err != nil {
		slog.Error("Error", "Errorstring", err)
		w.Write([]byte("Error while trying to read models files"))
	}

	bundle := []string{}

	for _, file := range files {
		if file.IsDir() {
			continue
		} else {
			bundle = append(bundle, file.Name())
		}
	}

	respBundle := ModelsResponse{
		Models: bundle,
	}

	json, err := json.Marshal(respBundle)

	w.WriteHeader(http.StatusOK)
	w.Write(json)

	return
}
