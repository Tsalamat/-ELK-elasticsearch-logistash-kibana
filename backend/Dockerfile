FROM golang:1.26-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/todo-api ./cmd

FROM alpine:3.20

WORKDIR /app
RUN adduser -D -H appuser && mkdir -p /app/logs && chown -R appuser:appuser /app

COPY --from=builder /out/todo-api /app/todo-api

USER appuser
EXPOSE 8080

CMD ["/app/todo-api"]
