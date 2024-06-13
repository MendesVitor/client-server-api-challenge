package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

type Response struct {
	Bid string `json:"bid"`
}

const (
	serverURL = "http://localhost:8080/cotacao"
	timeout   = 300 * time.Millisecond
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", serverURL, nil)
	if err != nil {
		log.Fatal("Error creating request:", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal("Error making request:", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Server returned non-200 status: %d %s", resp.StatusCode, resp.Status)
	}

	var result Response
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Fatal("Error decoding response:", err)
	}

	fileContent := fmt.Sprintf("Dólar: %s", result.Bid)
	if err := os.WriteFile("cotacao.txt", []byte(fileContent), 0644); err != nil {
		log.Fatal("Error writing to file:", err)
	}

	log.Println("Cotação salva em 'cotacao.txt'")
}
