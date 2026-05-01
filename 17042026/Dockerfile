# Build stage
FROM golang:1.22 AS builder
WORKDIR /app
COPY . .
RUN go mod init uvi || true
RUN go mod tidy
RUN go build -o app .

# Run stage
FROM debian:bookworm-slim
WORKDIR /app
COPY --from=builder /app/app .

CMD ["./app"]