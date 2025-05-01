FROM golang:1.23.6 AS builder

RUN apt-get update && apt-get install -y ca-certificates git-core ssh

WORKDIR /opt/load-balancer

# Данные приложения
COPY . .
COPY .docker/balancer/config.yaml ./config.yaml

# Установка переменных окружения
ENV TZ=Europe/Moscow

# Установка зависимостей
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Сборка
RUN go build -o app ./cmd/app/main.go
RUN chmod +x app

# Старт приложения
CMD ["./app"]