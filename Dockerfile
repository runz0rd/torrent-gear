FROM golang:alpine AS builder
RUN apk update && apk add --no-cache git ca-certificates

WORKDIR /build
COPY . .
RUN CGO_ENABLED=0 go build -o /gear cmd/main.go

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /gear /gear
ENTRYPOINT ["/gear"]