FROM golang:1.25 AS builder

WORKDIR /app

COPY go.mod go.sum ./


RUN go mod download

COPY . .

RUN go build -o driftguard ./backend

FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y ca-certificates

WORKDIR /app

COPY --from=builder /app/driftguard .
COPY --from=builder /app/backend/frontend ./frontend
EXPOSE 8080

CMD ["./driftguard"]