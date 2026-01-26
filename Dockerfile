# Stage 1: Build stage
FROM golang:1.24-alpine AS builder
RUN apk add --no-cache git

WORKDIR /src

COPY go.sum go.mod ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/svm-service ./cmd/main.go

# Stage 2: Final stage (Runtime)
FROM alpine:3.21.3

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/svm-service .

RUN chmod +x ./svm-service

EXPOSE 9924

CMD ["./svm-service"]