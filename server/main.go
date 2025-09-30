package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Cotacao struct {
	ID           int64     `json:"-" db:"id"`
	Valor        float64   `json:"bid" db:"valor"`
	DataConsulta time.Time `json:"create_date" db:"data_consulta"`
}

func InitDB() (*sql.DB, error) {
	// abre a conexão (cria arquivo caso não exista)
	db, err := sql.Open("sqlite3", "./banco.db")
	if err != nil {
		return nil, err
	}
	// testa a conexão
	if err = db.Ping(); err != nil {
		return nil, err
	}

	createTableSQL := `CREATE TABLE IF NOT EXISTS cotacoes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		valor REAL NOT NULL,
		data_consulta DATETIME NOT NULL
	); `

	// executa o SQL
	_, err = db.Exec(createTableSQL)
	if err != nil {
		return nil, err
	}

	fmt.Println("Banco de dados e tabela criados com sucesso!")
	return db, nil
}

func InsertCotacao(db *sql.DB, cotacao *Cotacao) error {
	stmt, err := db.Prepare("INSERT INTO cotacoes (valor, data_consulta) VALUES (?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(cotacao.Valor, cotacao.DataConsulta)
	if err != nil {
		return err
	}
	fmt.Println("cotação salva com sucesso!")
	return nil
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

func main() {
	cotacao, err := buscaCotacao()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Cotação: %f\n", cotacao.Valor)
	fmt.Printf("Data: %s\n", cotacao.DataConsulta)

	db, err := InitDB()
	if err != nil {
		panic(err)
	}

	err = InsertCotacao(db, cotacao)
	if err != nil {
		panic(err)
	}
}
