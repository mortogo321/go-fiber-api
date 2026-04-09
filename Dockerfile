# ---------- builder ----------
FROM golang:1.23-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /app/go-fiber-api .

# ---------- production ----------
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S appgroup \
    && adduser -S appuser -G appgroup

WORKDIR /app

COPY --from=builder /app/go-fiber-api .

RUN chown -R appuser:appgroup /app

USER appuser

EXPOSE 3000

ENTRYPOINT ["./go-fiber-api"]
