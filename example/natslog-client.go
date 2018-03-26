package main

import (
	"fmt"
	"io"
	"log"
	"time"

	stan "github.com/nats-io/go-nats-streaming"
)

func connectionCloser(c io.Closer) {
	if err := c.Close(); err != nil {
		log.Printf("close error: %s", err)
	}
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	const serviceName = "servicename"
	conn, err := stan.Connect(
		"test-cluster",
		"natslog-client",
		stan.NatsURL("nats://localhost:4222"),
	)
	if err != nil {
		return err
	}
	defer connectionCloser(conn)
	err = conn.Publish("natslog.subscribe", []byte(serviceName))
	if err != nil {
		return err
	}

	for i := 0; i < 10; i++ {
		log.Print("Publish")
		time.Sleep(time.Second)
		err := conn.Publish(serviceName, []byte(fmt.Sprintf("Hello world %d\n", i)))
		if err != nil {
			return err
		}
	}

	return nil
}
