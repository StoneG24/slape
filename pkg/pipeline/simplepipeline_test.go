package pipeline

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

func TestSetupEndpoint(t *testing.T) {
	// Setup test cases
	tests := []struct {
		name         string
		payload      string
		expectedCode int
		expectedMsg  string
	}{
		{
			name:         "successful response",
			payload:      `{"models": ["Dolphin3.0-Llama3.2-1B-Q4_K_M.gguf"]}`,
			expectedCode: 200,
		},
		{
			name:         "invalid JSON",
			payload:      `{"model": "Dolphin3.0-Llama3.2-1B-Q4_K_M.gguf"}`,
			expectedCode: 422,
		},
		{
			name:         "server error",
			payload:      `{"models": [".gguf"]}`,
			expectedCode: 500,
		},
	}

	// Common setup
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Create client with timeout
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// Test each case
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("POST", "http://localhost:"+port+"/simple/setup", strings.NewReader(tt.payload))
			if err != nil {
				t.Fatalf("Error creating request: %v", err)
			}

			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("Failed to send request: %v", err)
			}

			// Check response
			if resp.StatusCode != tt.expectedCode {
				t.Errorf("Unexpected status code: expected %d, got %d", tt.expectedCode, resp.StatusCode)
			}

			// Read response body
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("Failed to read response body: %v", err)
			}

			// Check response message if applicable
			if tt.expectedCode != 500 && strings.Contains(string(body), "error") {
				t.Errorf("Unexpected error message: %v", string(body))
			}
		})
	}
}

func TestShutdownEndpoint(t *testing.T) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://localhost:8080/simple/shutdown", nil)
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		return
	}

	errChan := make(chan error, 1)
	go func() {
		resp, err := client.Do(req)
		if err != nil {
			errChan <- fmt.Errorf("failed to execute request: %v", err)
			return
		}
		if resp.StatusCode != 500 {
			errChan <- fmt.Errorf("expected 500 status code, got %d", resp.StatusCode)
		}
		close(errChan)
	}()

	err = <-errChan
	if err != nil {
		fmt.Printf("Test failed: %v\n", err)
		return
	}

	fmt.Printf("Successfully tested shutdown endpoint\n")
}
