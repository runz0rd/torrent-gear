FROM golang:alpine AS builder
RUN apk update && apk add --no-cache git ca-certificates

WORKDIR /build
COPY . .
RUN CGO_ENABLED=0 go build -o /gear cmd/main.go

ENTRYPOINT ["/gear"]