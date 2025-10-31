package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"

	"github.com/google/uuid"
)

type BuyTicketRequest struct {
	Flight string `json:"flight"`
	Day    string `json:"day"`
	User   string `json:"user"`
}

type BuyTicketResponse struct {
	TransactionID string `json:"transactionID"`
}

type Ticket struct {
	TransactionID uuid.NullUUID `json:"transactionID"`
	FlightNumber  string        `json:"flight"`
	FlightDay     string        `json:"day"`
	Price         float64       `json:"price"`
	UserID        string        `json:"user"`
	Status        string        `json:"status"`
}

type FlightRequest struct {
	Flight string `json:"flight"`
	Day    string `json:"day"`
}

type FlightData struct {
	Flight string  `json:"flight"`
	Day    string  `json:"day"`
	Value  float64 `json:"value"`
}

type ExchangeToDolarResponse struct {
	Value float64 `json:"value"`
}

type SellRequest struct {
	Flight string `json:"flight"`
	Day    string `json:"day"`
}

type SellResponse struct {
	TransactionID string `json:"transactionID"`
}

type FidelityRequest struct {
	User  string `json:"user"`
	Bonus int    `json:"bonus"`
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

func GetFlight(flight string, day string) (*FlightData, error) {
	log.Printf("Iniciando busca por voo %s, dia %s", flight, day)

	endpoint := fmt.Sprintf("%s/flight?flight=%s&day=%s",
		cfg.URL.AirlinesHub,
		flight,
		day,
	)

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		log.Printf("ERRO: falha ao criar requisição para AirlinesHub: %v", err)
		return nil, fmt.Errorf("falha ao criar requisição para %s: %w", endpoint, err)
	}

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		log.Printf("ERRO: falha ao fazer requisição para AirlinesHub (%s): %v", endpoint, err)
		return nil, fmt.Errorf("falha ao fazer requisição para %s: %w", endpoint, err)
	}
	defer response.Body.Close()

	var flightData FlightData
	if err := json.NewDecoder(response.Body).Decode(&flightData); err != nil {
		log.Printf("ERRO: falha ao decodificar resposta do AirlinesHub: %v", err)
		return nil, fmt.Errorf("falha ao decodificar resposta de AirlinesHub: %w", err)
	}

	log.Printf("Sucesso: Voo encontrado: %+v", flightData)
	return &flightData, nil
}

func getDolarValueInReal() (float64, error) {
	log.Println("Iniciando busca por cotação do dólar")

	endpoint := fmt.Sprintf("%s/convert", cfg.URL.Exchange)
	req, err := http.NewRequest("GET", endpoint, nil)

	if err != nil {
		log.Printf("ERRO: falha ao criar requisição para Exchange: %v", err)
		return -1, fmt.Errorf("falha ao criar requisição para %s: %w", endpoint, err)
	}

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		log.Printf("ERRO: falha ao fazer requisição para Exchange (%s): %v", endpoint, err)
		return -1, fmt.Errorf("falha ao fazer requisição para %s: %w", endpoint, err)
	}
	defer response.Body.Close()

	var exchangeResponse ExchangeToDolarResponse
	if err := json.NewDecoder(response.Body).Decode(&exchangeResponse); err != nil {
		log.Printf("ERRO: falha ao decodificar resposta de Exchange: %v", err)
		return -1, fmt.Errorf("falha ao decodificar resposta de Exchange: %w", err)
	}

	log.Printf("Sucesso: Cotação do dólar obtida: %.2f", exchangeResponse.Value)
	return exchangeResponse.Value, nil
}

func RequestTicketSell(flight string, day string) (uuid.UUID, error) {
	log.Printf("Iniciando requisição de venda para voo %s, dia %s\n", flight, day)

	endpoint := fmt.Sprintf("%s/sell", cfg.URL.AirlinesHub)
	reqBody := SellRequest{
		Flight: flight,
		Day:    day,
	}

	reqData, err := json.Marshal(reqBody)
	if err != nil {
		log.Printf("ERRO: falha ao serializar request body: %v\n", err)
		return uuid.Nil, fmt.Errorf("falha ao serializar request body: %w", err)
	}

	resp, err := http.Post(endpoint, "application/json", bytes.NewBuffer([]byte(reqData)))
	if err != nil {
		log.Printf("ERRO: falha ao enviar requisição POST para %s: %v\n", endpoint, err)
		return uuid.Nil, fmt.Errorf("falha ao enviar requisição POST para %s: %w", endpoint, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		log.Printf("ERRO: servidor retornou status %d: %s\n", resp.StatusCode, string(bodyBytes))
		return uuid.Nil, fmt.Errorf("servidor retornou status não-OK %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var responsePayload SellResponse
	if err := json.NewDecoder(resp.Body).Decode(&responsePayload); err != nil {
		log.Printf("ERRO: falha ao decodificar resposta JSON: %v\n", err)
		return uuid.Nil, fmt.Errorf("falha ao decodificar resposta JSON: %w", err)
	}

	transactionUUID, err := uuid.Parse(responsePayload.TransactionID)
	if err != nil {
		log.Printf("ERRO: servidor retornou um transactionID inválido (%s): %v", responsePayload.TransactionID, err)
		return uuid.Nil, fmt.Errorf("servidor retornou um transactionID inválido: %w", err)
	}

	return transactionUUID, nil
}

func SendFidelityRequest(userID string, bonus int) (int, error) {
	log.Printf("Iniciando requisição de bônus para usuário %s, valor %d", userID, bonus)

	endpoint := fmt.Sprintf("%s/bonus", cfg.URL.Fidelity)
	reqBody := FidelityRequest{
		User:  userID,
		Bonus: bonus,
	}

	reqData, err := json.Marshal(reqBody)
	if err != nil {
		log.Printf("ERRO: falha ao serializar request body do fidelity: %v", err)
		return 0, fmt.Errorf("falha ao serializar request body: %w", err)
	}

	resp, err := http.Post(endpoint, "application/json", bytes.NewBuffer(reqData))
	if err != nil {
		log.Printf("ERRO: falha ao enviar requisição POST para fidelity (%s): %v", endpoint, err)
		return 0, fmt.Errorf("falha ao enviar requisição POST para %s: %w", endpoint, err)
	}

	defer resp.Body.Close()

	log.Printf("Serviço Fidelity respondeu com status: %d", resp.StatusCode)
	return resp.StatusCode, nil
}

func buyTicketHandler(w http.ResponseWriter, r *http.Request) {
	var body BuyTicketRequest
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		log.Printf("ERRO: JSON inválido recebido: %v", err)
		http.Error(w, "JSON inválido: "+err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("Requisição para /buyTicket recebida: %+v", body)

	if err := parseDate(body.Day); err != nil {
		log.Printf("ERRO: Data em formato inválido: %s", body.Day)
		http.Error(w, "Data em formato inválido", http.StatusBadRequest)
		return
	}

	log.Println("Buscando informações de voo em AirlinesHub...")

	flightData, err := GetFlight(body.Flight, body.Day)
	if err != nil {
		http.Error(w, "Erro na tentativa de buscar dados do voo no serviço AirlinesHub.", http.StatusInternalServerError)
		return
	}

	log.Printf("Voo buscado com sucesso. Dados de voo: %+v", flightData)

	log.Println("Buscando cotação do dolar em Exchange...")
	dolarExchangeRate, err := getDolarValueInReal()
	if err != nil {
		http.Error(w, "Erro na tentativa de buscar cotação do dolar no serviço Exchange.", http.StatusInternalServerError)
		return
	}

	price := dolarExchangeRate * flightData.Value

	ticket := Ticket{
		FlightNumber: body.Flight,
		FlightDay:    body.Day,
		Price:        price,
		UserID:       body.User,
		Status:       "PENDING_PAYMENT",
	}
	log.Printf("Ticket criado com sucesso %+v", ticket)

	transactionID, err := RequestTicketSell(ticket.FlightNumber, ticket.FlightDay)
	if err != nil {
		http.Error(w, "ERRO: falha ao realizar venda de ticket", http.StatusInternalServerError)
		return
	}

	ticket.TransactionID = uuid.NullUUID{UUID: transactionID, Valid: true}
	ticket.Status = "PAID"
	ticketDB[transactionID] = ticket
	log.Printf("Ticket armazenado no 'banco de dados' local: %+v", ticket)

	bonus := int(math.Round(flightData.Value))

	log.Printf("Enviando bônus de %d (baseado no valor US$ %.2f) para usuário %s", bonus, flightData.Value, body.User)

	statusCode, err := SendFidelityRequest(body.User, bonus)
	if err != nil {
		log.Printf("AVISO: Falha ao enviar bônus da venda %s: %v", transactionID.String(), err)
	} else {
		log.Printf("Sucesso: Bônus enviado para usuário %s", body.User)
	}

	response := BuyTicketResponse{
		TransactionID: transactionID.String(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}
