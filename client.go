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

// CotacaoResponse representa a resposta do servidor
type CotacaoResponse struct {
	Bid string `json:"bid"`
}

func main() {
	// Timeout de 300ms para receber o resultado do servidor
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost:8080/cotacao", nil)
	if err != nil {
		log.Fatalf("[CLIENT] Erro ao criar requisição: %v", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			log.Println("[CLIENT] Timeout: servidor excedeu 300ms")
		} else {
			log.Printf("[CLIENT] Erro na requisição ao servidor: %v", err)
		}
		return
	}
	defer resp.Body.Close()

	var cotacao CotacaoResponse
	if err := json.NewDecoder(resp.Body).Decode(&cotacao); err != nil {
		log.Printf("[CLIENT] Erro ao decodificar resposta: %v", err)
		return
	}

	// Salva a cotação no arquivo cotacao.txt
	content := fmt.Sprintf("Dólar: %s", cotacao.Bid)
	err = os.WriteFile("cotacao.txt", []byte(content), 0644)
	if err != nil {
		log.Printf("[CLIENT] Erro ao salvar arquivo: %v", err)
		return
	}

	fmt.Printf("[CLIENT] Cotação salva com sucesso: %s\n", content)
}
