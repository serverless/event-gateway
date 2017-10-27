FROM golang:1.9-alpine as builder
RUN apk add --update git

WORKDIR /go/src/github.com/serverless/event-gateway
COPY . .

RUN go get -u github.com/golang/dep/cmd/dep
RUN dep ensure
RUN go build -o event-gateway cmd/event-gateway/main.go

FROM alpine:3.6
WORKDIR /app/
COPY --from=builder /go/src/github.com/serverless/event-gateway/event-gateway .
EXPOSE 4000 4001
ENTRYPOINT ["./event-gateway"]