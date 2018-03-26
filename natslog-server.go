package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"

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
	log.Println("Started natslog")
	httpEnabled := getEnv("HTTP_ENABLED", "true")
	httpPort := getEnv("HTTP_PORT", "80")

	log.Println("Starting httpd server")
	if strings.EqualFold(httpEnabled, "true") {
		go http.ListenAndServe(":"+httpPort, http.FileServer(http.Dir("/var/log")))
	}
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

var lastProcessed map[string]uint64
var conn stan.Conn

func messageHandle(msg *stan.Msg) {
	if msg.Sequence > lastProcessed[msg.Subject] {
		fileflags := os.O_WRONLY | os.O_APPEND | os.O_CREATE
		f, err := os.OpenFile("/var/log/"+msg.Subject+".log", fileflags, 0660)
		if err != nil {
			log.Fatalf("error opening file: %v", err)
		}
		defer f.Close()
		_, ferr := f.Write(msg.Data)
		if ferr != nil {
			log.Fatalf("error writing file: %v", ferr)
		}
		lastProcessed[msg.Subject] = msg.Sequence
	}
	msg.Ack()
}

func registerNewHandle(msg *stan.Msg) {
	reg, err := regexp.Compile("[^a-zA-Z0-9]+")
	if err != nil {
		log.Fatal(err)
	}
	subject := reg.ReplaceAllString(string(msg.Data), "")
	natsDurableName := getEnv("NATS_DURABLE_NAME", "natslog-server") + subject
	log.Printf("Subscribing to new service %s %s", subject, natsDurableName)
	_, err = conn.Subscribe(
		subject,
		messageHandle,
		stan.DurableName(natsDurableName),
		stan.MaxInflight(1),
		stan.SetManualAckMode(),
	)
	if err != nil {
		log.Printf("error subscribing file: %v", err)
	}
	msg.Ack()
}

func run() error {
	lastProcessed = make(map[string]uint64)
	clusterName := getEnv("CLUSTER_NAME", "test-cluster")
	natsServer := getEnv("NATS_SERVER", "nats://localhost:4222")
	natsClientName := getEnv("NATS_CLIENT_NAME", "natslog-server")
	natsDurableName := getEnv("NATS_DURABLE_NAME", "natslog-server")
	natslogSubsribeSubject := getEnv("NATSLOG_SUBSCRIBE_SUBJECT", "natslog.subscribe")
	log.Printf("Connecting to %s", natsServer)
	c, err := stan.Connect(
		clusterName,
		natsClientName,
		stan.NatsURL(natsServer),
	)
	if err != nil {
		return err
	}
	conn = c
	defer connectionCloser(conn)

	sub, err := conn.Subscribe(
		natslogSubsribeSubject,
		registerNewHandle,
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
