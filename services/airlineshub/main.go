package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
)

var withOmissionFailure = false
var withTimeFailure = false

type Fail struct {
	Type        string
	Probability float64
	Duration    int
}

func (f Fail) makeOmissionFailure() {
	if withOmissionFailure == false {
		log.Println("[FAILURE] (Omissão) Iniciando estado de falha")
		withOmissionFailure = true
		go func() {
			log.Printf("[FAILURE] (Omissão) Sistema ficará em estado de falha por %d segundos", f.Duration)
			time.Sleep(time.Second * time.Duration(f.Duration))
			withOmissionFailure = false
			log.Println("[FAILURE] (Omissão) Encerrando estado de falha")
		}()
	}
}

func (f Fail) makeTimeFailure() {
	if withTimeFailure == false {
		log.Println("[FAILURE] (Time) Iniciando estado de falha")
		withTimeFailure = true
		go func() {
			log.Printf("[FAILURE] (Time) Sistema ficará em estado de falha por 10 segundos")
			time.Sleep(time.Second * time.Duration(10))
			withTimeFailure = false
			log.Println("[FAILURE] (Time) Encerrando estado de falha")
		}()
	}
}

type Flight struct {
	Code  string
	Day   string
	Value float64
}

type FlightRequest struct {
	Flight string `json:"flight"`
	Day    string `json:"day"`
}

type FlightResponse struct {
	Flight string  `json:"flight"`
	Day    string  `json:"day"`
	Value  float64 `json:"value"`
}

type SellResponse struct {
	TransactionID string `json:"transactionID"`
}

func main() {
	serviceName := "AirlinesHub"
	log.Printf("Iniciando serviço %s...", serviceName)
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthcheck", healthCheckHandler)
	mux.HandleFunc("GET /flight", flightHandler)
	mux.HandleFunc("POST /sell", sellHandler)

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

func generateRandomFlightValue() string {
	min := 100.0
	max := 250.0
	value := min + rand.Float64()*(max-min)
	return fmt.Sprintf("%.2f", value)
}

func flightHandler(w http.ResponseWriter, r *http.Request) {
	fail := Fail{
		Type:        "Omission",
		Probability: 0.2,
		Duration:    0,
	}
	if withOmissionFailure || rand.Float64() <= fail.Probability {
		log.Println("[FAILURE] Falha por omissão")
		fail.makeOmissionFailure()
		return
	}

	query := r.URL.Query()
	flightCode := query.Get("flight")
	flightDay := query.Get("day")

	if flightCode == "" {
		http.Error(w, "Falta parametro de busca: flight", http.StatusBadRequest)
		return
	}

	if flightDay == "" {
		http.Error(w, "Falta parametro de busca: day", http.StatusBadRequest)
		return
	}

	log.Printf("Valores recebidos - flight: %v, day: %v\n", flightCode, flightDay)

	flight := FlightRequest{
		Flight: flightCode,
		Day:    flightDay,
	}

	randValue, err := strconv.ParseFloat(generateRandomFlightValue(), 64)
	if err != nil {
		log.Fatalf("Erro ao converter string válida: %v", err)
	}

	log.Printf("Preço da passagem do voo: %v\n", randValue)

	f := Flight{
		Code:  flight.Flight,
		Day:   flight.Day,
		Value: randValue,
	}

	flightResponse := FlightResponse{
		Flight: f.Code,
		Day:    f.Day,
		Value:  f.Value,
	}

	log.Println("Retornando requisição com sucesso!")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(flightResponse)
}

func sellHandler(w http.ResponseWriter, r *http.Request) {

	fail := Fail{
		Type:        "Time",
		Probability: 0.1,
		Duration:    5,
	}

	if withTimeFailure || rand.Float64() <= fail.Probability {
		log.Println("[FAILURE] Falha por Time")
		fail.makeTimeFailure()
	}

	if withTimeFailure {
		log.Printf("[FAILURE] (Time) Paciência! O Sistema está lento!")

		for i := fail.Duration; i > 0; i-- {
			log.Printf("[FAILURE] (Time) Aguarde... %d", i)
			time.Sleep(1 * time.Second)
		}
	}

	var req FlightRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Printf("ERRO: Falha ao decodificar JSON do body: %v", err)
		http.Error(w, "JSON inválido: "+err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("Iniciando processo de venda: %+v", req)

	transactionID := uuid.New()

	resp := SellResponse{
		TransactionID: transactionID.String(),
	}

	log.Printf("Venda processada com sucesso. ID: %s", resp.TransactionID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("ERRO: Falha ao escrever resposta JSON: %v", err)
	}
}
