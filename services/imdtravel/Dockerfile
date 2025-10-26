FROM golang:1.25.3-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -a \
    -ldflags="-s -w" \
    -o /app/imdtravel-service \
    ./

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app/

COPY --from=builder /app/imdtravel-service .

RUN chmod +x ./imdtravel-service

EXPOSE 80
CMD ["./imdtravel-service"]
