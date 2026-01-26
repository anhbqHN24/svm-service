FROM golang:1.24-alpine as builder

WORKDIR /src
COPY go.sum go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /bin/svm-service .

FROM alpine:3.21.3

RUN mkdir /app

WORKDIR /app 
COPY --from=builder /bin/svm-service /app/svm-service
COPY settings settings
COPY bin bin

RUN chmod +x /app/svm-service bin/*
EXPOSE 9924
CMD ["bin/start_app.sh"]

