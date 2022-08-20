FROM golang:1.19-alpine as build-reflex

RUN go install github.com/cespare/reflex@latest

FROM golang:1.19-alpine
COPY --from=build-reflex /go/bin/reflex /go/bin/reflex

RUN apk update && apk upgrade && apk add bash git redis && rm -rf /var/cache/apk/*
WORKDIR /app

ENV GOOS=linux GOARCH=amd64 CGO_ENABLED=0

COPY . .
RUN go mod download

RUN go build -o /api cmd/main.go
