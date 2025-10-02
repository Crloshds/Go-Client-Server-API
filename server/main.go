package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
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

func InsertCotacao(ctx context.Context, db *sql.DB, cotacao *Cotacao) error {
	stmt, err := db.PrepareContext(ctx, "INSERT INTO cotacoes (valor, data_consulta) VALUES (?, ?)")
	if err != nil {
		return err
	}

	defer stmt.Close()
	_, err = stmt.ExecContext(ctx, cotacao.Valor, cotacao.DataConsulta)
	if err != nil {
		return err
	}

	fmt.Println("cotação salva com sucesso!")
	return nil
}

func buscaCotacao(ctx context.Context) (*Cotacao, error) {
	//cria request com context
	req, err := http.NewRequestWithContext(ctx, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	if err != nil {
		return nil, err
	}

	//executa request
	resp, err := http.DefaultClient.Do(req)
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

func BuscaCotacaoHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	if r.URL.Path != "/cotacao" {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	// Criar context com timeout de 200ms para a API de cotação
	ctxAPI, cancelAPI := context.WithTimeout(r.Context(), 200*time.Millisecond)
	defer cancelAPI()
	defer log.Println("Request finalizada")

	cotacao, err := buscaCotacao(ctxAPI)
	if err != nil {
		// Verificar se o timeout foi execido
		if ctxAPI.Err() == context.DeadlineExceeded {
			w.WriteHeader(http.StatusRequestTimeout)
			w.Write([]byte("Timeout ao buscar cotação: limite de 200ms excedido"))
			log.Println("Timeout ao buscar cotação: limite de 200ms excedido")
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Erro interno ao buscar cotação"))
			log.Println("Erro interno ao buscar cotação")
		}
		return
	}

	ctxDB, cancelDB := context.WithTimeout(r.Context(), 10*time.Millisecond)
	defer cancelDB()

	if err := InsertCotacao(ctxDB, db, cotacao); err != nil {
		if ctxDB.Err() == context.DeadlineExceeded {
			log.Println("Timeout de 10ms atingido ao tentar salvar no banco")
		} else {
			log.Println("Erro ao salvar no banco:", err)
		}
	}

	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	log.Println("Request processada com sucesso.")
	json.NewEncoder(w).Encode(cotacao.Valor)
}

func main() {

	db, err := InitDB()
	if err != nil {
		panic(err)
	}
	defer db.Close()

	fmt.Println("Servidor iniciado na porta 8080")
	http.HandleFunc("/cotacao", func(w http.ResponseWriter, r *http.Request) {
		BuscaCotacaoHandler(w, r, db)
	})
	http.ListenAndServe(":8080", nil)
}
