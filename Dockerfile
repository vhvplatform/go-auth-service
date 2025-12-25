# Build stage
FROM golang:1.25.5-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git

# Copy go.work and modules
COPY go.work go.work
COPY pkg/go.mod pkg/go.sum pkg/
COPY services/auth-service/go.mod services/auth-service/go.sum services/auth-service/

# Download dependencies
WORKDIR /app/services/auth-service
RUN go mod download

# Copy source code
WORKDIR /app
COPY pkg/ pkg/
COPY services/auth-service/ services/auth-service/

# Build the application
WORKDIR /app/services/auth-service
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/bin/auth-service ./cmd/main.go

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/bin/auth-service .

# Copy .env.example as reference
COPY .env.example .env.example

# Expose ports
EXPOSE 50051 8081

CMD ["./auth-service"]
