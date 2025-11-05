package main

/*
O IMDTravel envia um request para o Exchange, via GET para o endpoint /convert, sem
parâmetros.
A resposta deve ser um número real positivo que indica a taxa de conversão da moeda
(assumindo que precisa converter de dólar para real). Gere esse valor de forma randômica
com variação entre 1/5 e 1/6 (ou seja, 1 dólar pode variar entre 5 e 6 reais).

*/
import (
	"encoding/json"
	"errors"
	"log"
	"math/rand"
	"net/http"
	"time"
)

type ExchangeToDolarResponse struct {
	Value float64 `json:"value"`
}

var withFailure = false

type Fail struct {
	Type        string
	Probability float64
	Duration    int
}

func (f Fail) makeFailure() error {
	if withFailure == false {
		log.Println("[FAILURE] Iniciando estado de falha")
		withFailure = true
		go func() {
			log.Printf("[FAILURE] Sistema ficará em estado de falha por %d segundos", f.Duration)
			time.Sleep(time.Second * time.Duration(f.Duration))
			withFailure = false
			log.Println("[FAILURE] Encerrando estado de falha")
		}()
	}

	return errors.New("falha ao tentar buscar valor do dolar")
}

func main() {
	serviceName := "Exchange"
	log.Printf("Iniciando serviço %s...", serviceName)
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthcheck", healthCheckHandler)
	mux.HandleFunc("GET /convert", conversionToDolar)

	port := ":80"
	log.Printf("Serviço %s rodando na porta %s", serviceName, port[1:])
	log.Fatal(http.ListenAndServe(port, mux))
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := struct {
		Message string `json:"message"`
	}{
		Message: "OK",
	}
	json.NewEncoder(w).Encode(response)
}

func conversionToDolar(w http.ResponseWriter, r *http.Request) {
	rate, err := getDolarRatePrice()
	if err != nil {
		errMsg := struct {
			Message string `json:"message"`
		}{
			Message: err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(errMsg)
		return
	}

	rateDolarResponse := ExchangeToDolarResponse{
		Value: rate,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(rateDolarResponse)

}

func getDolarRatePrice() (float64, error) {
	fail := Fail{
		Type:        "Error",
		Probability: 0.1,
		Duration:    5,
	}

	if withFailure || rand.Float64() <= fail.Probability {
		log.Println("[FAILURE] Falha por erro")

		err := fail.makeFailure()
		return -1, err
	}

	min := 5.0
	max := 6.0

	value := min + rand.Float64()*(max-min)

	return value, nil
}
