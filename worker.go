package main

import (
	_ "expvar"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
)

// Catch os signals and process them.
func signalDispatcher() {
	// Make a channel for signals.
	ch := make(chan os.Signal, 1)
	defer close(ch)

	// We want catch only SIGINT and SIGKILL.
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	for range ch {
		queue.Stop()
	}
}

const appVersion = "0.0.2"

// Retry in sec.
var timeout = [...]int{1, 3, 5, 15, 30, 60, 60, 60, 60}
var port *int
var configDir *string
var queueFile *string
var queueSize *int
var queue *Queue
var actions ActionsMap

func main() {
	log.Printf("[INFO] Elastica Worker! (%s)", appVersion)
	log.Printf("[INFO] PID: %d", os.Getpid())

	// Parse arguments
	port = flag.Int("port", 8312, "Listen port")
	configDir = flag.String("config", "/etc/worker", "Path with configuration (/etc/worker)")
	queueFile = flag.String("save", "/tmp/worker.state", "File to save queue (/tmp/worker.state)")
	queueSize = flag.Int("queue", 10000, "Queue size")
	flag.Parse()

	// Load configuration
	actions = NewActionsMap(*configDir)

	// Make queue object
	queue = NewQueue(*queueSize, *queueFile)

	// Run queue events process
	go queue.Process()

	// Catch SIGINT and SIGTERM
	go signalDispatcher()

	// Run HTTP server
	new(HttpServer).Run(*port)
}
