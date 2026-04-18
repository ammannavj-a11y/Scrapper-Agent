FROM golang:1.22-alpine AS builder
RUN apk add --no-cache ca-certificates git
WORKDIR /app
COPY go.mod ./
RUN go mod download 2>/dev/null || go mod tidy
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /service ./cmd/

FROM alpine:3.19
RUN apk add --no-cache ca-certificates
COPY --from=builder /service /service
EXPOSE 8080
ENTRYPOINT ["/service"]
