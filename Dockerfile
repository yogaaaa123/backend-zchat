# Stage 1: Build
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Cache Go modules
COPY go.mod go.sum ./
RUN go mod download

# Build
COPY . .
RUN CGO_ENABLED=0 go build -o /app/server ./cmd/server/main.go

# Stage 2: Run
FROM alpine:3.19

RUN apk add --no-cache ca-certificates && \
    addgroup -S app && adduser -S app -G app

WORKDIR /app

COPY --from=builder /app/server .

USER app

EXPOSE 8080

CMD ["./server"]
