# syntax=docker/dockerfile:1

FROM golang:1.23 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . ./

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/avito-shop ./cmd/app

FROM alpine:3.16

WORKDIR /root/
COPY --from=builder /app/avito-shop .
EXPOSE 8080

CMD ["./avito-shop"]
