# Projeto Tolerancia a Falhas Unidade 2

## Integrantes do Grupo
- Felipe Barauna Costa Sousa (felipe.barauna.131@ufrn.edu.br)
- Ivison Santana Cau Filho (ivisoncaufilho@gmail.com)

## Pré-requisitos

Para rodar este projeto, você precisará ter o [Docker](https://www.docker.com/) e o Docker Compose instalados em sua máquina.

## Como Rodar o Projeto

1.  Clone o repositório.
2. Copie o arquivo .env.example e renomeie-o para .env (não é necessário modificá-lo)

Exemplo:

```bash
cp .env.example .env

```

3.  Na pasta `services/` execute:

```bash
docker compose up --build
```

Para rodar em segundo plano:

```bash
docker compose up --build -d
  ```


## Endpoints

### IMDTravel

GET http://localhost:8080/healthcheck

Response:
```json
{"message":"OK"}
```

POST http://localhost:8080/buyTicket (Rota principal)

Payload:

```json
{
    "flight": "05A8EF14",
    "day": "2025-12-01",
    "user": "joao",
    "ft": true
}
```

Use o campo 'ft' do payload para dizer se a requisição deve utilizar das técnicas de tolerância a Falhas
implementadas ou não.

Response:
```json
{"transactionID":"019a2220-9ff6-7d85-9cbd-7ffd84639366"}
```

### AirlinesHub

GET http://localhost:8081/flight

Query Params:

- flight (string)

- day (string)

Response:
```json
{"flight": "05B7EF14","day": "2025-08-12", "value": 105.20}
```

Example:

GET http://localhost:8081/flight?flight="05A8EF14"&day="2025-12-25"

Response:
```json
{"flight":"05A8EF14","day":"2025-12-25","value":207.35}
```

POST http://localhost:8081/sell

Payload:
```json
{
    "flight": "05A8EF14",
    "day": "2025-12-01",
}
```

Response:
```json
{"transactionID":"019a2220-9ff6-7d85-9cbd-7ffd84639366"}
```

### Exchange

GET http://localhost:8082/convert

Response:
```json
{"value": 5.3}
```

Example:

GET http://localhost:8082/convert

Response:
```json
{"value":5.9}
```

### Fidelity

GET http://localhost:8083/healthcheck

Response: 
```json
{"message": "OK"}
```

POST http://localhost:8083/bonus

Payload:
```json
{
  "user": "user-123",
  "bonus": 305
}
```

Response:
```json
{"message": "Bônus registrado com sucesso"}
```

Example:

POST http://localhost:8083/bonus

Payload:
```json
{
  "user": "user123",
  "bonus": 75
}
```

Response:
```json
{"message": "Bônus registrado com sucesso"}
```
