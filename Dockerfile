ROM golang:1.24.2-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /ride-notification ./cmd/notification-service/main.go

FROM alpine:3.18
WORKDIR /app

COPY --from=builder /ride-notification /app/ride-notification
COPY .env /app/.env
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/


RUN addgroup -S appgroup && adduser -S appuser -G appgroup \
    # Create the log directory and give permission to appuser
    && mkdir -p /app/log \
    && chown -R appuser:appgroup /app

USER appuser

HEALTHCHECK --interval=30s --timeout=3s CMD wget -qO- http://localhost:8080/health || exit 1

EXPOSE 8080
CMD ["/app/ride-notification"]