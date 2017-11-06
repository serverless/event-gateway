FROM golang:1.9-alpine as builder
RUN apk add --update git

WORKDIR /go/src/github.com/serverless/event-gateway
COPY . .

RUN go get -u github.com/hashicorp/go-plugin
RUN go get -u github.com/hashicorp/go-hclog
RUN go get -u golang.org/x/net/context
RUN go get -u golang.org/x/net/http2
RUN go get -u golang.org/x/net/trace
RUN go get -u github.com/golang/dep/cmd/dep
RUN dep ensure
RUN go build -o event-gateway cmd/event-gateway/main.go

FROM alpine:3.6
RUN apk add --no-cache ca-certificates
WORKDIR /app/
COPY --from=builder /go/src/github.com/serverless/event-gateway/event-gateway .
EXPOSE 4000 4001
ENTRYPOINT ["./event-gateway"]