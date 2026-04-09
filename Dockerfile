# ---- Build Stage ----
FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o server .

# ---- Runtime Stage ----
FROM alpine:3.20

RUN apk --no-cache add ca-certificates && \
    adduser -D appuser

WORKDIR /app

COPY --from=builder /app/server .

USER appuser

EXPOSE 3000

CMD ["./server"]
