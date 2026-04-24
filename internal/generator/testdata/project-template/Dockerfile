FROM golang:1.25.0 AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/__PROJECT_NAME__ ./cmd/main.go

FROM alpine:3.21

RUN apk --no-cache add ca-certificates tzdata && \
    addgroup -S appgroup && \
    adduser -S appuser -G appgroup

WORKDIR /app
COPY --from=builder /bin/__PROJECT_NAME__ /app/bin/__PROJECT_NAME__
COPY --from=builder /app/configs /app/configs
COPY --from=builder /app/internal/docs /app/internal/docs
COPY --from=builder /app/migrations /app/migrations
RUN mkdir -p /app/logs && chown -R appuser:appgroup /app

USER appuser
CMD ["/app/bin/__PROJECT_NAME__", "serve"]
