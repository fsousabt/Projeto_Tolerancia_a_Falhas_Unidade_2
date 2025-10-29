FROM golang:1.25.3-alpine AS builder

WORKDIR /app

COPY go.mod ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -a \
    -ldflags="-s -w" \
    -o /app/fidelity-service \
    ./

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app/

COPY --from=builder /app/fidelity-service .

RUN chmod +x /app/fidelity-service

EXPOSE 80
CMD ["./fidelity-service"]
