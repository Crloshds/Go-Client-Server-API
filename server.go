package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// AwesomeAPIResponse representa a resposta da API externa
type AwesomeAPIResponse struct {
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

// CotacaoResponse representa a resposta JSON do servidor para o cliente
type CotacaoResponse struct {
	Bid string `json:"bid"`
}

var db *sql.DB

func initDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "cotacoes.db")
	if err != nil {
		return nil, err
	}

	createTableSQL := `CREATE TABLE IF NOT EXISTS cotacoes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		bid TEXT NOT NULL,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	_, err = db.Exec(createTableSQL)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func saveCotacao(ctx context.Context, bid string) error {
	query := "INSERT INTO cotacoes (bid) VALUES ($1)"
	_, err := db.ExecContext(ctx, query, bid)
	return err
}

func fetchCotacao(ctx context.Context) (*AwesomeAPIResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var apiResp AwesomeAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, err
	}

	return &apiResp, nil
}

func cotacaoHandler(w http.ResponseWriter, r *http.Request) {
	// Timeout de 200ms para chamar a API externa
	ctxAPI, cancelAPI := context.WithTimeout(r.Context(), 200*time.Millisecond)
	defer cancelAPI()

	apiResp, err := fetchCotacao(ctxAPI)
	if err != nil {
		if ctxAPI.Err() == context.DeadlineExceeded {
			log.Println("[SERVER] Timeout: API externa excedeu 200ms")
		} else {
			log.Printf("[SERVER] Erro ao buscar cotação na API externa: %v", err)
		}
		http.Error(w, `{"error": "erro ao obter cotação da API externa"}`, http.StatusInternalServerError)
		return
	}

	bid := apiResp.USDBRL.Bid

	// Timeout de 10ms para persistir no banco de dados
	ctxDB, cancelDB := context.WithTimeout(r.Context(), 10*time.Millisecond)
	defer cancelDB()

	err = saveCotacao(ctxDB, bid)
	if err != nil {
		if ctxDB.Err() == context.DeadlineExceeded {
			log.Println("[SERVER] Timeout: persistência no banco excedeu 10ms")
		} else {
			log.Printf("[SERVER] Erro ao persistir cotação no banco: %v", err)
		}
		// Continua mesmo com erro de banco, retorna a cotação ao cliente
	}

	response := CotacaoResponse{Bid: bid}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func main() {
	var err error
	db, err = initDB()
	if err != nil {
		log.Fatalf("Erro ao inicializar banco de dados: %v", err)
	}
	defer db.Close()

	http.HandleFunc("/cotacao", cotacaoHandler)

	fmt.Println("Servidor iniciado na porta 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
