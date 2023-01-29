package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type DollarExchange struct {
	Bid string `json:"bid"`
}

const REQUEST_TIMEOUT = time.Millisecond * 300

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), REQUEST_TIMEOUT)

	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost:8080/cotacao", nil)

	if err != nil {
		panic(err)
	}

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Println("Ocorreu um erro ao consultar a API")
		return
	}

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		panic(err)
	}

	var dollarExchange DollarExchange

	err = json.Unmarshal(body, &dollarExchange)

	if err != nil {
		panic(err)
	}

	var file *os.File

	file, err = os.Create("cotacao.txt")

	if err != nil {
		panic(err)
	}

	defer file.Close()

	_, err = file.WriteString(fmt.Sprintf("Dólar: %s\n", dollarExchange.Bid))

	if err != nil {
		panic(err)
	}
}
