FROM golang:1.8-alpine

RUN apk add --update curl git
RUN curl https://glide.sh/get | sh

WORKDIR $GOPATH/src/github.com/serverless/event-gateway
COPY . .

RUN glide install
RUN go build

ENTRYPOINT ["./event-gateway"]