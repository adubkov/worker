package main

import (
	_ "expvar"
	"flag"
	"log"
	"os"
	"os/signal"
)

// Catch os signals and process them.
func signalDispatcher() {
	// Make a channel for signals.
	ch := make(chan os.Signal, 1)
	defer close(ch)

	// We want catch only SIGINT and SIGKILL.
	signal.Notify(ch, os.Interrupt, os.Kill)
	for range ch {
		queue.Stop()
	}
}

// Retry in sec.
var timeout = [...]int{1, 3, 5, 15, 30, 60, 300, 900, 3600, 3600, 3600, 86400}
var port *int = flag.Int("port", 8312, "Listen port")
var configDir *string = flag.String("config", "/etc/worker", "Path with configuration /etc/worker")
var queueFile *string = flag.String("save", "/tmp/worker.state", "File to save queue (/tmp/worker.state")
var queueSize *int = flag.Int("queue", 10000, "Queue size")
var queue *Queue
var actions ActionsMap

func main() {
	log.Println("[INFO] Elastica Worker!")
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
