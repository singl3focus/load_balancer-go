FROM golang:1.23-alpine

WORKDIR /opt/test-server
COPY pkg/test-server/server.go .

RUN go build -o server server.go

CMD ["./server"]