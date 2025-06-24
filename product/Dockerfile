# Build stage
FROM golang:tip-20250620-alpine3.21 AS builder
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the Go app
RUN go build -o main .

# Run stage
FROM debian:bookworm-slim
WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/main .

# Expose port (change if your app uses a different port)
EXPOSE 8080

# Run the binary
CMD ["./main"]