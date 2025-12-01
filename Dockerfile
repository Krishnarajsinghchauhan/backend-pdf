FROM golang:1.22 AS builder

WORKDIR /app
COPY . .
RUN go build -o server ./cmd/api

FROM debian:bookworm
WORKDIR /app

COPY --from=builder /app/server .
COPY .env /app/.env

EXPOSE 8080
CMD ["./server"]
