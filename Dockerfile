FROM golang:1.11-alpine3.8 as builder
RUN apk add --update git
RUN apk add ca-certificates

WORKDIR /app
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags '-w -s' -a -installsuffix cgo -o event-gateway cmd/event-gateway/main.go

FROM scratch
WORKDIR /
COPY --from=builder /app/event-gateway /
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
EXPOSE 4000 4001
ENTRYPOINT ["/event-gateway"]
