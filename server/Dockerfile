# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN go build -o wowa-server .

# Runtime stage
FROM alpine:latest

WORKDIR /root/
COPY --from=builder /app/wowa-server .

CMD ["./wowa-server"]
