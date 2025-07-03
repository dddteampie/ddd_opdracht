# Build stage
FROM golang:1.24.4-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the Go application
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/main .

FROM alpine:latest

RUN apt-get update && apt-get install -y curl
WORKDIR /app

COPY --from=builder /app/main .

EXPOSE 8084

CMD ["/app/main"]
