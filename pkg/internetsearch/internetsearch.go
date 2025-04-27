package internetsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"sort"
	"strings"

	"github.com/gocolly/colly"
	"github.com/openai/openai-go"
)

type (
	EmbeddingRequest struct {
		Prompt []string `json:"prompt"`
	}

	EmbeddingResponse struct {
		Response openai.CreateEmbeddingResponse
	}

	VectorList struct {
		Points   []Point
		Elements []string
		index    int
	}

	Point struct {
		ID     int
		Vector []float64
	}

	Neighbor struct {
		Point    Point
		Distance float64
	}
)

const (
	embedmodel = "snowflake-arctic-embed-l-v2.0-q4_k_m.gguf"
	//Typical values:
	// > 0.9 → very close (same idea, rephrased)
	// 0.7–0.9 → somewhat related
	// < 0.7 → probably not related
	similarityThreshold = 0.50
)

// InternetSearch is used to search the internet with an models query request
//
// First the model should generate a query.
// Then the search tool should make a request to a privacy oriented search engine.
// Once this is done it should web scrape the top five websites, and store it in the contex box.
// Internet search should not be compared with the rest of tools because it can be
// dangerous if not used properly, hence why it is serperate.
func InternetSearch(ctx context.Context, query string) VectorList {

	collyCollector := colly.NewCollector()

	data := []Point{}
	elements := []string{}
	index := 0

	vecs := VectorList{data, elements, index}

	query = strings.ReplaceAll(query, " ", "+")
	query = strings.ReplaceAll(query, "\n", "")
	queryurl := fmt.Sprintf("https://html.duckduckgo.com/html/?q=%s", query)

	log.Println(queryurl)

	linkIndex := 0
	maxlinks := 3

	// Find and visit all links
	// gather all of the elements
	collyCollector.OnHTML(".result__url", func(element *colly.HTMLElement) {
		//counter used to limit the number of websites
		//searches for links on duckduckgo and creates links to individual sites to be scraped
		link := "https://" + strings.TrimSpace(element.Text)

		if linkIndex < maxlinks {
			vecs.scrape(ctx, link)
			linkIndex++
		}
	})

	err := collyCollector.Visit(queryurl)
	if err != nil {
		log.Println("error while scraping duckduckgo", err)
	}

	// send the vecs to embedding
	embeddings := vecs.embedGuy()

	// add them to the vecs
	vecs.addGuy(embeddings)

	return vecs
}

// used to scrape individual sites
func (v *VectorList) scrape(ctx context.Context, link string) {

	collyCollector := colly.NewCollector()

	//scrapes all paragraph elements from each webpage
	collyCollector.OnHTML("p", func(element *colly.HTMLElement) {
		//concatenates all text from paragraph elements
		// add the guy to the element list
		log.Println(element.Text)
		v.Elements = append(v.Elements, element.Text)
		v.index++
	})
	collyCollector.OnHTML("code", func(element *colly.HTMLElement) {
		log.Println(element.Text)
		v.Elements = append(v.Elements, element.Text)
		v.index++
	})
	collyCollector.OnRequest(func(req *colly.Request) {
		log.Println("Visiting", req.URL)
	})

	//start scraping by visiting the page
	err := collyCollector.Visit(link)
	if err != nil {
		log.Println("error while scraping webpage")
	}
}

func (v *VectorList) embedGuy() []openai.Embedding {

	var chunkedText []string
	text := v.Elements[:v.index]
	bulkText := strings.Join(text, " ")
	v.index = 0
	for index := 0; index < len(bulkText); index += 200 {
		if index+250 > (len(bulkText) - 1) {
			chunkedText = append(chunkedText, bulkText[index:])
			v.index++
		} else {
			chunkedText = append(chunkedText, bulkText[index:index+250])
			v.index++
		}
	}
	v.Elements = chunkedText
	text = chunkedText
	var prompt EmbeddingRequest
	prompt.Prompt = text
	embeddingPrompt, err := json.Marshal(prompt)
	if err != nil {
		log.Println("The fucking json Marshal didn't fucking work you STUPID MOTHERFUCKER")
	}
	resp, err := http.Post("http://localhost:8080/emb/generate",
		http.DetectContentType(embeddingPrompt),
		bytes.NewReader(embeddingPrompt),
	)
	if err != nil {
		log.Println("Sorry Vito what was that? You cut out.")
	}
	defer resp.Body.Close()

	var embeddingResponse EmbeddingResponse

	err = json.NewDecoder(resp.Body).Decode(&embeddingResponse)
	if err != nil {
		log.Println("Failed to Read response body in internet search")
	}

	vectorizedText := embeddingResponse.Response.Data

	return vectorizedText
}

func (v *VectorList) addGuy(embeddings []openai.Embedding) {

	for i := range len(embeddings) {
		// add point to the array of data
		point := Point{i, embeddings[i].Embedding}
		log.Println(point)

		v.Points = append(v.Points, point)
	}
}

func normalize(vec []float64) []float64 {
	var norm float64
	for _, v := range vec {
		norm += v * v
	}
	norm = math.Sqrt(norm)
	for i := range vec {
		vec[i] /= norm
	}
	return vec
}

func cosineSimilarity(vec1, vec2 []float64) float64 {
	var dot, normA, normB float64
	for i := range len(vec1) {
		dot += vec1[i] * vec2[i]
		normA += vec1[i] * vec1[i]
		normB += vec2[i] * vec2[i]
	}

	if normA == 0 || normB == 0 {
		return 0
	}
	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}

func KnnSearch(data []Point, query []float64, k int) []Neighbor {
	var neighbors []Neighbor
	normquery := normalize(query)
	for _, point := range data {
		normpoint := normalize(point.Vector)
		dist := cosineSimilarity(normquery, normpoint)
		log.Println("Simularity Score", dist)
		if dist >= similarityThreshold {
			neighbors = append(neighbors, Neighbor{Point: point, Distance: dist})
		}
	}

	sort.Slice(neighbors, func(i, j int) bool {
		return neighbors[i].Distance < neighbors[j].Distance
	})

	if len(neighbors) < k {
		return neighbors
	}
	return neighbors[:k]
}
