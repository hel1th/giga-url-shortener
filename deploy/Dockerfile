# syntax=docker/dockerfile:1

# ---------------------------------------------------------
# Build Stage
# ---------------------------------------------------------
FROM golang:alpine AS builder

WORKDIR /app

# Install ca-certificates and tzdata for secure HTTPS and timezone handling
RUN apk add --no-cache ca-certificates tzdata

# Install goose migration tool
RUN CGO_ENABLED=0 go install github.com/pressly/goose/v3/cmd/goose@v3.24.1

# Download Go module dependencies (layer cached)
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build application binaries
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/link-service ./cmd/link-service
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/analytics-service ./cmd/analytics-service

# ---------------------------------------------------------
# Migrations Target
# ---------------------------------------------------------
FROM alpine:3.21 AS migrations

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /go/bin/goose /bin/goose

WORKDIR /migrations
COPY migrations/ /migrations/

ENTRYPOINT ["/bin/goose"]

# ---------------------------------------------------------
# Link Service Target
# ---------------------------------------------------------
FROM alpine:3.21 AS link-service

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app
COPY --from=builder /bin/link-service /app/link-service

EXPOSE 8080

ENTRYPOINT ["/app/link-service"]

# ---------------------------------------------------------
# Analytics Service Target
# ---------------------------------------------------------
FROM alpine:3.21 AS analytics-service

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app
COPY --from=builder /bin/analytics-service /app/analytics-service

EXPOSE 8082 50051

ENTRYPOINT ["/app/analytics-service"]

# ---------------------------------------------------------
# Default Target
# ---------------------------------------------------------
FROM link-service
