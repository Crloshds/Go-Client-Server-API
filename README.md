# Client-Server API - Cotação USD/BRL

Sistema interligado em Go composto por um **servidor HTTP** e um **cliente**, que consomem a API de câmbio AwesomeAPI, persistem dados em SQLite e trocam informações respeitando limites estritos de tempo (timeout) via `context`.

---

##  Estrutura do Projeto

```
client-server-api/
├── server.go      # Servidor HTTP (porta 8080)
├── client.go      # Cliente HTTP
├── go.mod         # Módulo Go (gerado automaticamente)
├── go.sum         # Checksum de dependências (gerado automaticamente)
├── cotacoes.db    # Banco SQLite (gerado na primeira execução)
└── cotacao.txt    # Arquivo com cotação salva (gerado pelo cliente)
```

---

## Como Executar

### 1. Inicializar o Módulo Go

Na pasta do projeto, execute:

```bash
go mod init client-server-api
go mod tidy
```

> O `go mod tidy` baixa automaticamente a dependência `github.com/mattn/go-sqlite3`.

---

### 2. Iniciar o Servidor

```bash
go run server.go
```

Saída esperada:
```
Servidor iniciado na porta 8080
```

O servidor ficará escutando na porta **8080**, pronto para receber requisições.

---

### 3. Executar o Cliente

Em **outro terminal**, na mesma pasta:

```bash
go run client.go
```

Saída esperada:
```
[CLIENT] Cotação salva com sucesso: Dólar: 5.7276
```

O arquivo `cotacao.txt` será criado com o formato:
```
Dólar: 5.7276
```

---

## Timeouts

| Componente | Timeout | Descrição |
|------------|---------|-----------|
| **Server → API Externa** | 200ms | Tempo máximo para consultar a API AwesomeAPI |
| **Server → Banco SQLite** | 10ms | Tempo máximo para persistir a cotação no banco |
| **Client → Server** | 300ms | Tempo máximo para receber resposta do servidor local |

Todos os timeouts são implementados com o pacote **`context`** do Go.

---

## Arquitetura

```
┌─────────────┐     300ms      ┌─────────────┐     200ms      ┌──────────────────┐
│   client.go │ ──────────────→ │  server.go  │ ──────────────→│ AwesomeAPI       │
│  (Cliente)  │   (timeout)    │  (Servidor) │   (timeout)    │ (USD-BRL)        │
│             │                │             │                │                  │
│  Recebe     │                │  Recebe     │                │  Retorna JSON    │
│  campo bid  │                │  JSON da    │                │  com cotação     │
│  e salva    │                │  API e      │                │                  │
│  em         │                │  persiste   │                │                  │
│  cotacao.txt│                │  no SQLite  │                │                  │
│             │                │  (10ms)     │                │                  │
└─────────────┘                └─────────────┘                └──────────────────┘
                                      │
                                      │ 10ms (timeout)
                                      ▼
                               ┌─────────────┐
                               │  SQLite     │
                               │  cotacoes.db│
                               └─────────────┘
```

---

##  Banco de Dados

O servidor cria automaticamente o arquivo `cotacoes.db` (SQLite) com a seguinte tabela:

```sql
CREATE TABLE IF NOT EXISTS cotacoes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    bid TEXT NOT NULL,
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

---

## Endpoints

| Método | Endpoint | Descrição |
|--------|----------|-----------|
| `GET` | `/cotacao` | Retorna a cotação atual do USD/BRL em formato JSON (`{"bid": "5.7276"}`) |

---

## Logs

### Servidor (`server.go`)
- Caso o timeout da API externa (200ms) seja excedido:
  ```
  [SERVER] Timeout: API externa excedeu 200ms
  ```
- Caso o timeout do banco (10ms) seja excedido:
  ```
  [SERVER] Timeout: persistência no banco excedeu 10ms
  ```

### Cliente (`client.go`)
- Caso o timeout do servidor (300ms) seja excedido:
  ```
  [CLIENT] Timeout: servidor excedeu 300ms
  ```

---

## Tecnologias Utilizadas

- **Go** — Linguagem principal
- **`net/http`** — Servidor e cliente HTTP
- **`context`** — Controle de timeouts
- **`database/sql`** — Abstração de banco de dados
- **`github.com/mattn/go-sqlite3`** — Driver SQLite para Go
- **AwesomeAPI** — API pública de câmbio (`https://economia.awesomeapi.com.br`)

---

