package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync/atomic"

	stan "github.com/nats-io/go-nats-streaming"
)

func getEnv(key, fallback string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		value = fallback
	}
	return value
}

func connectionCloser(c io.Closer) {
	if err := c.Close(); err != nil {
		log.Printf("close error: %s", err)
	}
}

func main() {
	httpEnabled := getEnv("HTTP_ENABLED", "true")
	httpPort := getEnv("HTTP_PORT", "80")

	if strings.EqualFold(httpEnabled, "true") {
		http.ListenAndServe(":"+httpPort, http.FileServer(http.Dir("/var/log")))
	}
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

var lastProcessed uint64

func messageHandle(msg *stan.Msg) {
	if msg.Sequence > lastProcessed {
		fileflags := os.O_WRONLY | os.O_APPEND | os.O_CREATE
		f, err := os.OpenFile(msg.Subject+".log", fileflags, 0660)
		if err != nil {
			log.Fatalf("error opening file: %v", err)
		}
		defer f.Close()
		_, ferr := f.Write(msg.Data)
		if ferr != nil {
			log.Fatalf("error writing file: %v", ferr)
		}
		atomic.SwapUint64(&lastProcessed, msg.Sequence)
	}
	msg.Ack()
}

func run() error {
	clusterName := getEnv("CLUSTER_NAME", "test-cluster")
	natsServer := getEnv("NATS_SERVER", "nats://localhost:4222")
	natsClientName := getEnv("NATS_CLIENT_NAME", "natslog-server")
	natsDurableName := getEnv("NATS_DURABLE_NAME", "natslog-server")
	natsSubject := getEnv("NATS_SUBJECT", "natslog")

	conn, err := stan.Connect(
		clusterName,
		natsClientName,
		stan.NatsURL(natsServer),
	)
	if err != nil {
		return err
	}
	defer connectionCloser(conn)

	sub, err := conn.Subscribe(
		natsSubject,
		messageHandle,
		stan.DurableName(natsDurableName),
		stan.MaxInflight(1),
		stan.SetManualAckMode(),
	)
	if err != nil {
		return err
	}
	defer connectionCloser(sub)
	select {}
}
