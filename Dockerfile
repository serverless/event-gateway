FROM golang:1.8-alpine as builder

RUN apk add --update curl git
RUN curl https://glide.sh/get | sh

WORKDIR /go/src/github.com/serverless/event-gateway
COPY . .

RUN glide install
RUN go build -o event-gateway cmd/event-gateway/main.go

FROM alpine:3.6
WORKDIR /app/
COPY --from=builder /go/src/github.com/serverless/event-gateway/event-gateway .
ENTRYPOINT ["./event-gateway"]