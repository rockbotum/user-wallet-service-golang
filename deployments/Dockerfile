# deployments/Dockerfile
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Кэшируем зависимость
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Собираем статический бинарник
RUN CGO_ENABLED=0 GOOS=linux go build -o /user-wallet-service ./cmd/app

# Финальный образ
FROM alpine:3.20

RUN apk add --no-cache ca-certificates

# Непривилегированный пользователь для запуска приложения
RUN addgroup -S app && adduser -S -G app app

WORKDIR /app

COPY --from=builder /user-wallet-service .

USER app

EXPOSE 8080

CMD ["./user-wallet-service"]