package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Dollar struct {
	Cotacao string `json:"bid"`
	Data    string `json:"create_date"`
}

func main() {

	dolar, err := buscaCotacao()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Cotação: R$ %s\n", dolar.Cotacao)
	fmt.Printf("Data: %s\n", dolar.Data)
}

func buscaCotacao() (*Dollar, error) {
	resp, err := http.Get("https://economia.awesomeapi.com.br/json/last/USD-BRL")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]Dollar
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	cotacao := result["USDBRL"]

	return &cotacao, nil
}
