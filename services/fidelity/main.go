package main

import (
	"encoding/json"
	"log"
	"net/http"
)

type BonusRequest struct {
	User  string `json:"user"`
	Bonus int    `json:"bonus"`
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
