package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type ExchangeRate struct {
	USDBRL struct {
		Code       string `json:"code"`
		Codein     string `json:"codein"`
		Name       string `json:"name"`
		High       string `json:"high"`
		Low        string `json:"low"`
		VarBid     string `json:"varBid"`
		PctChange  string `json:"pctChange"`
		Bid        string `json:"bid"`
		Ask        string `json:"ask"`
		Timestamp  string `json:"timestamp"`
		CreateDate string `json:"create_date"`
	} `json:"USDBRL"`
}

const (
	apiURL        = "https://economia.awesomeapi.com.br/json/last/USD-BRL"
	dbPath        = "./exchange_rate.db"
	serverAddress = ":8080"
)

func fetchDollarRate(ctx context.Context) (*ExchangeRate, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var rate ExchangeRate

	err = json.Unmarshal(body, &rate)
	if err != nil {
		return nil, err
	}

	return &rate, nil
}

func saveRateToDB(ctx context.Context, rate string) error {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.ExecContext(ctx, "CREATE TABLE IF NOT EXISTS exchange_rate (id INTEGER PRIMARY KEY, rate TEXT, timestamp DATETIME DEFAULT CURRENT_TIMESTAMP)")
	if err != nil {
		return err
	}

	_, err = db.ExecContext(ctx, "INSERT INTO exchange_rate (rate) VALUES (?)", rate)
	return err
}

func exchangeRateHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 200*time.Millisecond)
	defer cancel()

	rate, err := fetchDollarRate(ctx)
	if err != nil {
		http.Error(w, "Failed to fetch dollar rate", http.StatusInternalServerError)
		log.Println("Error fetching dollar rate:", err)
		return
	}

	dbCtx, dbCancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer dbCancel()

	if err := saveRateToDB(dbCtx, rate.USDBRL.Bid); err != nil {
		http.Error(w, "Failed to save rate to database", http.StatusInternalServerError)
		log.Println("Error saving rate to database:", err)
		return
	}

	response := map[string]string{"bid": rate.USDBRL.Bid}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		log.Println("Error encoding response:", err)
		return
	}
}

func main() {
	http.HandleFunc("/cotacao", exchangeRateHandler)
	log.Println("Server running on port", serverAddress)
	if err := http.ListenAndServe(serverAddress, nil); err != nil {
		log.Panic("Server failed to start:", err)
	}
}
