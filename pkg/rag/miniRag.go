/*
Pacakge rag is for working with rag solutions in order to aid in context generation for our slms.
These solutions should be easy to use on low end hardware since this is intended for consummer grade hardware.
*/
package rag

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

// The current plan is to implement, MiniRag, https://github.com/HKUDS/MiniRAG.
// This would allow for fast parsing of documents that could aid in getting slms
// more instruction on how to perform their tasking.

// func Insert() {

// }

func Retrieve(query string, mode string) string {
	fullQuery := strings.NewReader(fmt.Sprintf("{'query': %s, 'mode': %s, ''}", query, mode))
	resp, err := http.Post("http://localhost:9721/query", "application/json", fullQuery)
	if err != nil {
		fmt.Printf("Kill yourself idk %s", "-Vito\n")
	}
	defer resp.Body.Close()
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Vito Gay for this %s", "one")
	}

	// fmt.Printf("Status Code: %s\nResponse: %s", resp.Status, responseBody)

	if resp.StatusCode == http.StatusOK {
		return string(responseBody)
	} else {
		fmt.Printf("Error in Retrieval: %s\n", resp.Status)
		return "None"
	}

}

// func Delete() {

// }
