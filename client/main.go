package client

import (
	"io"
	"net/http"
	"os"
	"time"
)

type Dollar struct {
	Cotacao float64   `json:"bid"`
	Data    time.Time `json:"create_date"`
}

func main() {
	req, err := http.Get("https://economia.awesomeapi.com.br/json/last/USD-BRL")
	if err != nil {
		panic(err)
	}
	defer req.Body.Close()

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()
	io.Copy(os.Stdout, res.Body)
}
