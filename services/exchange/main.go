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
	"log"
	"math/rand"
	"net/http"
)

type ExchangeToDolarResponse struct {
	Value float64 `json:"value"`
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

	rate := generateRandomRateValue()

	rateDolarResponse := ExchangeToDolarResponse{
		Value: rate,
	}

	w.Header().Set("Contentt-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(rateDolarResponse)

}

func generateRandomRateValue() float64 {
	min := 5.0
	max := 6.0

	value := min + rand.Float64()*(max-min)

	return value
}
