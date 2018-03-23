# natslog
A lightweight log aggregator using NATS

![natslog](https://raw.githubusercontent.com/mgumowsk/natslog/naster/natslog.png)

# Setup
Start a NATS Streaming instance using [docker nats-streaming image](https://hub.docker.com/_/nats-streaming/)
```
$ docker run -d -p 4222:4222 -name nats-streaming nats-streaming
```

Start natslog server
```
$ docker run -d natslog-server
$ docker run -d -e NATS_SERVER=nats://nats-streaming:4222 -p 80:80 --link nats-streaming:nats-streaming mgumowsk/natslog
```
the events are written to /var/log/natslog.log


Use client to log events on natslog server
```
$ go run natslog-client.go
```

# Environment Variables
    CLUSTER_NAME
NATS streaming cluster name, defaults to "test-cluster"

    NATS_SERVER

NATS server URL, default is "nats://localhost:4222"

    NATS_CLIENT_NAME
    
Client name for natslog server, default "natslog-server"

    NATS_DURABLE_NAME
    
NATS streaming durable name, default "natslog-server"

    NATS_SUBJECT
    
NATS message subject and log filename, default "natslog"

    HTTP_ENABLED
    
Enable optional static web server for /var/log folder, default "true"

    HTTP_PORT

default "80"
