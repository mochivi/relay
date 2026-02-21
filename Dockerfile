# Build stage
FROM golang:1.24-alpine AS builder
WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o relay ./cmd/relay

# Run stage
FROM alpine:3.20
RUN apk --no-cache add ca-certificates
WORKDIR /app

COPY --from=builder /build/relay /app/relay

# Default config path (override with CONFIG_PATH or mount configs)
ENV CONFIG_PATH=/app/configs/docker.yaml
COPY configs/docker.yaml /app/configs/docker.yaml

EXPOSE 8080
ENTRYPOINT ["/app/relay"]
