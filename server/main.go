package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

type Cotacao struct {
	ID           int64     `json:"-" db:"id"`
	Valor        float64   `json:"bid" db:"valor"`
	DataConsulta time.Time `json:"create_date" db:"data_consulta"`
}

func main() {

	dolar, err := buscaCotacao()
	if err != nil {
		panic(err)
	}
	fmt.Println(dolar)
	fmt.Printf("Cotação: %f\n", dolar.Valor)
	fmt.Printf("Data: %s\n", dolar.DataConsulta)
}

func buscaCotacao() (*Cotacao, error) {
	resp, err := http.Get("https://economia.awesomeapi.com.br/json/last/USD-BRL")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var cotacao Cotacao
	if err := json.Unmarshal(body, &cotacao); err != nil {
		return nil, err
	}

	return &cotacao, nil

}

// Quando o json.Unmarshal encontra uma struct com este método, ele chama essa função em vez do parsing padrão.
func (q *Cotacao) UnmarshalJSON(data []byte) error {

	//Salva em um struct temporária
	var raw struct {
		USDBRL struct {
			Bid        string `json:"bid"`
			CreateDate string `json:"create_date"`
		} `json:"USDBRL"`
	}

	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("erro ao fazer unmarshal do JSON: %w", err)
	}

	bid, err := strconv.ParseFloat(raw.USDBRL.Bid, 64)
	if err != nil {
		return fmt.Errorf("erro ao converter valor para float64: %w", err)
	}

	createDate, err := time.Parse("2006-01-02 15:04:05", raw.USDBRL.CreateDate)
	if err != nil {
		return fmt.Errorf("erro ao converter data: %w", err)
	}

	q.Valor = bid
	q.DataConsulta = createDate

	return nil
}
