FROM golang:alpine AS goose-builder
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go install github.com/pressly/goose/v3/cmd/goose@v3.24.1


FROM golang:alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download


COPY . .


RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/link-service ./cmd/link-service

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/analytics-service ./cmd/analytics-service



FROM alpine:3.21 AS migrations
RUN apk add --no-cache ca-certificates tzdata
COPY --from=goose-builder /go/bin/goose /bin/goose
WORKDIR /migrations
COPY migrations/ /migrations/
ENTRYPOINT ["/bin/goose"]


FROM alpine:3.21 AS link-service
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /bin/link-service /app/link-service
EXPOSE 8080
ENTRYPOINT ["/app/link-service"]


FROM alpine:3.21 AS analytics-service
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /bin/analytics-service /app/analytics-service
EXPOSE 8082 50051
ENTRYPOINT ["/app/analytics-service"]


FROM link-service