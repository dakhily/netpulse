FROM golang:1.25-alpine AS builder 
WORKDIR /build 
COPY . .
RUN go mod download 
ENV CGO_ENABLED=0 GOOS=linux
RUN go build -o ./netpulse

FROM alpine:3.18
WORKDIR /app
COPY  --from=builder /build/netpulse ./netpulse
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
CMD ["/app/netpulse"]

