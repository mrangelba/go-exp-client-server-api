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

type UsdBrl struct {
	Code      string `json:"code"`
	Codein    string `json:"codein"`
	Name      string `json:"name"`
	High      string `json:"high"`
	Low       string `json:"low"`
	VarBid    string `json:"varBid"`
	PctChange string `json:"pctChange"`
	Bid       string `json:"bid"`
}

type AwesomeapiUsdBrlDto struct {
	UsdBrl UsdBrl `json:"USDBRL"`
}

type DollarExchangeDto struct {
	Bid string `json:"bid"`
}

const GET_TIMEOUT = time.Millisecond * 200

const SAVE_TIMEOUT = time.Millisecond * 10

func main() {
	createDatabase()

	mux := http.NewServeMux()

	mux.HandleFunc("/cotacao", getDollarExchangeHandler)

	http.ListenAndServe(":8080", mux)
}

func getDollarExchangeHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), GET_TIMEOUT)

	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Println(err)
		return
	}

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Println(err)
		return
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Println(err)
		return
	}

	var awesomeapiUsdBrl AwesomeapiUsdBrlDto

	err = json.Unmarshal(body, &awesomeapiUsdBrl)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Println(err)
		return
	}

	var bid = awesomeapiUsdBrl.UsdBrl.Bid

	err = saveDollarExchange(bid)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Println(err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(&DollarExchangeDto{
		Bid: bid,
	})

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Println(err)
		return
	}
}

func createDatabase() {
	db, err := sql.Open("sqlite3", "./data.db")

	if err != nil {
		panic(err)
	}

	defer db.Close()

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS exchange (id INTEGER PRIMARY KEY AUTOINCREMENT, bid VARCHAR(10))")

	if err != nil {
		panic(err)
	}
}

func saveDollarExchange(value string) error {
	ctx, cancel := context.WithTimeout(context.Background(), SAVE_TIMEOUT)

	defer cancel()

	db, err := sql.Open("sqlite3", "./data.db")

	if err != nil {
		return err
	}

	defer db.Close()

	tx, err := db.BeginTx(ctx, nil)

	defer tx.Rollback()

	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, "INSERT INTO exchange (bid) VALUES (?)", value)

	if err != nil {
		return err
	}

	return tx.Commit()
}
