# Build stage
FROM golang:1.25.5-alpine AS builder

WORKDIR /app

# Install required tools
RUN apk add --no-cache git

# Copy go.mod and go.sum for dependency caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the entire source code
COPY . .

# Build the service
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o /app/bin/auth-service ./cmd/main.go

# Final stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
RUN addgroup -S appgroup && adduser -S appuser -u 1000 -G appgroup
WORKDIR /app
RUN chown 1000:1000 /app
COPY --from=builder --chown=1000:1000 /app/bin/auth-service .
USER 1000
EXPOSE 50051 8081
CMD ["./auth-service"]
