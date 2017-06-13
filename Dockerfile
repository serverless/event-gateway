FROM golang:1.8.1-alpine

ADD ./bin/gateway /usr/bin/gateway
ENTRYPOINT ["/usr/bin/gateway", "-embed-master", "-embed-cli-addr=http://0.0.0.0:2379"]
