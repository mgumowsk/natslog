package main

import (
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
	conn, err := stan.Connect(
		"test-cluster",
		"natslog-client",
		stan.NatsURL("nats://localhost:4222"),
	)
	if err != nil {
		return err
	}
	defer connectionCloser(conn)

	for i := 0; i < 10; i++ {
		log.Print("Publish")
		time.Sleep(time.Second)
		err := conn.Publish("natslog", []byte("Hello world\n"))
		if err != nil {
			return err
		}
	}

	return nil
}
