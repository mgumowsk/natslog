FROM golang as builder

RUN go get github.com/nats-io/go-nats-streaming
WORKDIR /go/src/app
ADD natslog-server.go /go/src/app/natslog-server.go
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main /go/src/app/natslog-server.go

FROM alpine
RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*
WORKDIR /root
COPY --from=builder /go/src/app/main /usr/bin/natslog
EXPOSE 80

ENTRYPOINT ["/usr/bin/natslog"]
