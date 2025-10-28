package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
)

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

func main() {
	serviceName := "AirlinesHub"
	log.Printf("Iniciando serviço %s...", serviceName)
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthcheck", healthCheckHandler)
	mux.HandleFunc("GET /flight", flightHandler)

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

	flight := FlightRequest{
		Flight: flightCode,
		Day:    flightDay,
	}

	randValue, err := strconv.ParseFloat(generateRandomFlightValue(), 64)
	if err != nil {
		log.Fatalf("Erro ao converter string válida: %v", err)
	}

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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(flightResponse)
}
