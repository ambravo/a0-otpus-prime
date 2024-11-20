# Dockerfile
FROM golang:1.23.3-alpine AS builder

# Install git and SSL certificates
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bot cmd/server/main.go

# Create final image
FROM alpine:3.20.3

WORKDIR /app

# Copy necessary files from builder
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/bot .
COPY internal/assets/templates ./internal/web/templates

# Create non-root user
RUN adduser -D -g '' appuser && \
    chown -R appuser:appuser /app

USER appuser

EXPOSE 8080

CMD ["./bot"]