FROM golang:1.21.4-alpine as builder

WORKDIR /usr/src/app

COPY . .
RUN go mod download && go mod verify

RUN CGO_ENABLED=0 go build -o /usr/local/bin/realtime-notification cmd/realtime/main.go

EXPOSE 8090-8091