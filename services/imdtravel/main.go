// imdtravel/main.go
package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
)

var httpClient = &http.Client{
	Timeout: 2 * time.Second,
}

var flightCache = make(map[string]*FlightData)

var dolarCache []float64 // guarda últimas cotações

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

type APIError struct {
	StatusCode int    `json:"statusCode"`
	Message    string `json:"message"`
}

type FaultToleranceConfig struct {
	Ft string `json:"ft"`
}

func newAPIError(statusCode int, err error) *APIError {
	return &APIError{
		StatusCode: statusCode,
		Message:    err.Error(),
	}
}

func writeError(w http.ResponseWriter, err *APIError) {
	writeJSON(w, err.StatusCode, map[string]string{"error": err.Message})
}

func writeJSON(w http.ResponseWriter, code int, payload any) {
	response, err := json.Marshal(payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("erro interno ao codificar JSON"))
		log.Printf("Erro ao fazer marshal do JSON: %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

var ErrTicketSellTimeout = errors.New("Timeout excedido no processo de venda de passagem aérea.")

func cacheKey(flight, day string) string {
	return flight + "|" + day
}

func avg(values []float64) float64 {
	if len(values) == 0 {
		return -1 // "não tem cache"
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
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

// Função De Retry
// T = tipo de retorno (aqui será *FlightData)
// attempts = quantas tentativas
// fn = função a ser chamada
func retry[T any](attempts int, fn func() (T, error)) (T, error) {
	var zero T
	var err error

	for i := 0; i < attempts; i++ {
		var result T
		result, err = fn()
		if err == nil {
			return result, nil
		}

		// backoff exponencial simples
		sleep := time.Duration(math.Pow(2, float64(i))) * 200 * time.Millisecond
		log.Printf("Tentativa %d falhou: %v — retry em %v", i+1, err, sleep)
		time.Sleep(sleep)
	}

	return zero, err
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

	response, err := httpClient.Do(req)
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
	flightCache[cacheKey(flight, day)] = &flightData
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

	if response.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(response.Body)

		var errMsg struct {
			Message string `json:"message"`
		}

		if err := json.Unmarshal(bodyBytes, &errMsg); err == nil && errMsg.Message != "" {
			log.Printf("ERRO: Serviço Exchange retornou status %d: %s", response.StatusCode, errMsg.Message)
			return -1, fmt.Errorf("serviço Exchange falhou, %s", errMsg.Message)
		}

		log.Printf("ERRO: Serviço Exchange retornou status não-OK %d: %s", response.StatusCode, string(bodyBytes))
		return -1, fmt.Errorf("serviço Exchange retornou status não-OK %d", response.StatusCode)
	}

	var exchangeResponse ExchangeToDolarResponse
	if err := json.NewDecoder(response.Body).Decode(&exchangeResponse); err != nil {
		log.Printf("ERRO: falha ao decodificar resposta de Exchange: %v", err)
		return -1, fmt.Errorf("falha ao decodificar resposta de Exchange: %w", err)
	}

	log.Printf("Sucesso: Cotação do dólar obtida: %.2f", exchangeResponse.Value)

	// Atualiza cache mantendo os últimos 10 valores
	dolarCache = append(dolarCache, exchangeResponse.Value)
	if len(dolarCache) > 10 {
		dolarCache = dolarCache[len(dolarCache)-10:]
	}

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

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer([]byte(reqData)))
	if err != nil {
		log.Printf("ERRO: falha ao montar requisição POST para %s: %v\n", endpoint, err)
		return uuid.Nil, fmt.Errorf("falha ao montar requisição POST para %s: %w", endpoint, err)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		var netErr net.Error
		if errors.As(err, &netErr) && netErr.Timeout() {
			log.Println("Falha na requisição devido a alta latência. Timeout de 2s excedido.")
			return uuid.Nil, ErrTicketSellTimeout
		}
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
		apiErr := newAPIError(http.StatusBadRequest, fmt.Errorf("JSON inválido: %w", err))
		writeError(w, apiErr)
		return
	}

	log.Printf("Requisição para /buyTicket recebida: %+v", body)

	if err := parseDate(body.Day); err != nil {
		log.Printf("ERRO: Data em formato inválido: %s", body.Day)
		apiErr := newAPIError(http.StatusBadRequest, fmt.Errorf("data em formato inválido: %s", body.Day))
		writeError(w, apiErr)
		return
	}
	log.Println("Buscando informações de voo em AirlinesHub...")

	// ---- RETRY + TIMEOUT AQUI ----
	flightData, err := retry[*FlightData](3, func() (*FlightData, error) {
		return GetFlight(body.Flight, body.Day)
	})
	if err != nil {
		key := cacheKey(body.Flight, body.Day)
		cached, ok := flightCache[key]

		if ok {
			log.Printf("AVISO: Falha ao buscar voo online, usando valor do cache: %+v", cached)
			flightData = cached
		} else {
			log.Printf("ERRO: falha ao buscar dados do voo após retries e sem cache: %v", err)
			apiErr := newAPIError(http.StatusInternalServerError, fmt.Errorf("erro ao buscar dados do voo: %w", err))
			writeError(w, apiErr)
			return
		}
	}

	log.Printf("Voo buscado com sucesso. Dados de voo: %+v", flightData)

	log.Println("Buscando cotação do dolar em Exchange...")
	dolarExchangeRate, err := getDolarValueInReal()
	if err != nil {
		log.Printf("AVISO: falha ao buscar cotação do dólar: %v", err)

		// tenta usar média das últimas 10 taxas
		media := avg(dolarCache)

		if media > 0 {
			log.Printf("Usando média das últimas %d cotações: %.2f", len(dolarCache), media)
			dolarExchangeRate = media
		} else {
			// sem cache -> erro real
			log.Printf("ERRO: nenhum valor de cache disponível para cotação do dólar")
			apiErr := newAPIError(http.StatusInternalServerError, fmt.Errorf("falha ao buscar cotação do dólar e cache vazio"))
			writeError(w, apiErr)
			return
		}
	}

	price := dolarExchangeRate * flightData.Value

	price, _ = strconv.ParseFloat(fmt.Sprintf("%.2f", price), 64)

	log.Printf("Valor convertido para real com sucesso: %.2f", price)

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
		if errors.Is(err, ErrTicketSellTimeout) {
			log.Printf("[ERRO] (Falha Graciosa) " + err.Error())
			ticket.Status = "FAILED"
			log.Printf("[ERRO] Pagamento da passagem aérea não será processado!")
			apiErr := newAPIError(http.StatusGatewayTimeout, fmt.Errorf("falha ao realizar venda de ticket: %w", err))
			writeError(w, apiErr)
			return
		}
		log.Printf("ERRO: falha ao realizar venda de ticket: %v", err)
		apiErr := newAPIError(http.StatusInternalServerError, fmt.Errorf("falha ao realizar venda de ticket: %w", err))
		writeError(w, apiErr)
		return
	}

	ticket.TransactionID = uuid.NullUUID{UUID: transactionID, Valid: true}
	ticket.Status = "PAID"
	ticketDB[transactionID] = ticket
	log.Printf("Ticket armazenado no 'banco de dados' local: %+v", ticket)

	bonus := int(math.Round(flightData.Value))

	log.Printf("Enviando bônus de %d (baseado no valor US$ %.2f) para usuário %s", bonus, flightData.Value, body.User)

	_, err = SendFidelityRequest(body.User, bonus)
	if err != nil {
		log.Printf("AVISO: Falha ao enviar bônus da venda %s: %v", transactionID.String(), err)
	} else {
		log.Printf("Sucesso: Bônus enviado para usuário %s", body.User)
	}

	response := BuyTicketResponse{
		TransactionID: transactionID.String(),
	}

	log.Printf("Retornando ID da transação: %s", transactionID.String())

	writeJSON(w, http.StatusOK, response)
}
