package main

import (
	"log"
	"os"
)

type URL struct {
	AirlinesHub string
	Exchange    string
	Fidelity    string
}
type Config struct {
	URL
}

const (
	AIRLINES_HUB_URL = "AIRLINES_HUB_URL"
	EXCHANGE_URL     = "EXCHANGE_URL"
	FIDELITY_URL     = "FIDELITY_URL"
)

func MakeConfig() Config {
	airlinesHubURL := os.Getenv(AIRLINES_HUB_URL)
	exchangeURL := os.Getenv(EXCHANGE_URL)
	fidelityURL := os.Getenv(FIDELITY_URL)

	if airlinesHubURL == "" {
		log.Fatalf("Faltando variável de ambiente %s", AIRLINES_HUB_URL)
	}

	if exchangeURL == "" {
		log.Println("Faltando variável de ambiente %s", EXCHANGE_URL)
	}

	if fidelityURL == "" {
		log.Fatalf("Faltando variável de ambiente %s", FIDELITY_URL)
	}

	var cfg = Config{
		URL: URL{
			AirlinesHub: airlinesHubURL,
			Exchange:    exchangeURL,
			Fidelity:    fidelityURL,
		},
	}

	log.Printf("config: %+v", cfg)

	return cfg
}

var cfg = MakeConfig()

func GetConfig() Config {
	return cfg
}
