FROM golang:1.8-alpine

RUN apk add --update curl git
RUN curl https://glide.sh/get | sh

WORKDIR $GOPATH/src/github.com/serverless/event-gateway
COPY . .

RUN glide install
RUN go build -o event-gateway cmd/event-gateway/main.go

ENTRYPOINT ["./event-gateway"]