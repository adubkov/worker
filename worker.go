package main

import (
	_ "expvar"
	"log"
	"os"
	"os/signal"
)

func signalDispatcher() {
	ch := make(chan os.Signal, 1)
	defer close(ch)
	signal.Notify(ch, os.Interrupt, os.Kill)
	for range ch {
		queue.Stop()
	}
}

// Will make retry in sec
var timeout = [...]int{1, 5, 15, 30, 60, 180, 3600, 86400}
var port string = "8080"
var configDir string = "./config"
var queueFile string = "/tmp/worker.state"
var queueSize int = 10000

var queue *Queue
var actions ActionsMap

func main() {
	log.Println("[INFO] Elastica Worker!")

	// Load configuration
	actions = NewActionsMap(configDir)

    // Make queue object
	queue = NewQueue(queueSize, queueFile)

	// Run queue events process
	go queue.Process()

	// Catch SIGINT and SIGTERM
	go signalDispatcher()

	// Run HTTP server
	new(HttpServer).Run(port)
}
