# ---------- build stage ----------
FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" \
    -o /mcp-service ./cmd/server

# ---------- final stage ----------
FROM alpine:3.19

RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

COPY --from=builder /mcp-service /usr/local/bin/mcp-service

EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/mcp-service"]
