FROM golang:1.9-alpine as builder
RUN apk add --update git
RUN apk add ca-certificates

WORKDIR /go/src/github.com/serverless/event-gateway
COPY . .

RUN go get -u github.com/hashicorp/go-plugin
RUN go get -u github.com/hashicorp/go-hclog
RUN go get -u golang.org/x/net/context
RUN go get -u golang.org/x/net/http2
RUN go get -u golang.org/x/net/trace
RUN go get -u github.com/golang/dep/cmd/dep
RUN dep ensure
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags '-w -s' -a -installsuffix cgo -o event-gateway cmd/event-gateway/main.go

FROM scratch
WORKDIR /
COPY --from=builder /go/src/github.com/serverless/event-gateway/event-gateway /
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
EXPOSE 4000 4001
ENTRYPOINT ["/event-gateway"]
