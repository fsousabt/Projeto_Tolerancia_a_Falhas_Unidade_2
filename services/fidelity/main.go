package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"os"
)

type BonusRequest struct {
	User  string `json:"user"`
	Bonus int    `json:"bonus"`
}

type Fail struct {
	Type        string
	Probability float64
}

func main() {
	serviceName := "Fidelity"
	log.Printf("Iniciando serviço %s...", serviceName)
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthcheck", healthCheckHandler)
	mux.HandleFunc("POST /bonus", bonusHandler)

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

func bonusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}
	//Simulação de falha - Crash - Request 4|Crash (Stop)| 0.02 | indefinido
	fail := Fail{
		Type:        "Crash",
		Probability: 0.02,
	}

	if rand.Float64() < fail.Probability {
		log.Println("[FAILURE] Falha por Crash - Encerrando serviço")
		os.Exit(1) //Processo encerrado
	}

	var req BonusRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Erro ao decodificar JSON", http.StatusBadRequest)
		return
	}

	log.Printf("Recebido bônus para usuário %s: %d pontos", req.User, req.Bonus)

	w.WriteHeader(http.StatusOK)
	response := struct {
		Message string `json:"message"`
	}{
		Message: "Bônus registrado com sucesso",
	}
	json.NewEncoder(w).Encode(response)
}
