package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/google/uuid"
)

type BuyTicketRequest struct {
	Flight string `json:"flight"`
	Day    string `json:"day"`
	User   string `json:"user"`
}

type Ticket struct {
	TransactionID uuid.UUID `json:"transactionID"`
	FlightNumber  string    `json:"flight"`
	FlightDay     string    `json:"day"`
	UserID        string    `json:"user"`
	Status        string    `json:"status"`
}

type BuyTicketResponse struct {
	TransactionID string `json:"transactionID"`
}

type FlightRequest struct {
	Flight string `json:"flight"`
	Day    string `json:"day"`
}

type FlightData struct {
	Flight string `json:"flight"`
	Day    string `json:"day"`
	Value  string `json:"value"`
}

var ticketDB = make(map[uuid.UUID]Ticket)

func main() {
	log.Println("Iniciando serviço IMDTravel...")
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthcheck", healthCheckHandler)
	mux.HandleFunc("POST /buyTicket", buyTicketHandler)

	port := ":80"
	log.Printf("Serviço IMDTravel rodando na porta %s", port[1:])
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

func parseDate(day string) error {
	return nil
}

func buyTicketHandler(w http.ResponseWriter, r *http.Request) {
	var body BuyTicketRequest
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		http.Error(w, "JSON inválido: "+err.Error(), http.StatusBadRequest)
		return
	}

	transactionID, err := uuid.NewV7()

	if err != nil {
		http.Error(w, "Erro ao gerar ID de transação", http.StatusInternalServerError)
	}

	err = parseDate(body.Day)

	if err != nil {
		http.Error(w, "Data em formato inválido", http.StatusBadRequest)
	}

	ticket := Ticket{
		TransactionID: transactionID,
		FlightNumber:  body.Flight,
		FlightDay:     body.Day,
		UserID:        body.User,
		Status:        "PENDING_PAYMENT",
	}

	log.Printf("Ticket criado com sucesso %+v\n", ticket)

	ticketDB[ticket.TransactionID] = ticket

	log.Printf("Ticket armazenado no banco de dados, ID: %s", ticket.TransactionID.String())

	response := BuyTicketResponse{
		TransactionID: transactionID.String(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}
