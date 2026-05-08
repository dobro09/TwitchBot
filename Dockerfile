# Стадия сборки
FROM golang:1.26-bookworm AS builder

RUN apt-get update && apt-get install -y git ca-certificates

WORKDIR /app

# Кэшируем зависимости
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходники и собираем статический бинарник
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /app/twitchbot main.go

# Финальный образ
FROM alpine:3.21

RUN apk add --no-cache ca-certificates

WORKDIR /app

# Копируем бинарник и миграции
COPY --from=builder /app/twitchbot .
COPY db/migrations /app/db/migrations

ENTRYPOINT ["/app/twitchbot"]