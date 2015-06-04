package main

import (
	_ "expvar"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
)

func saveToFile(file string, data string) error {
	f, e := os.Create(file)
	defer f.Close()
	if e != nil {
		return e
	}
	f.WriteString(data)
	return e
}

func loadFromFile(file string, data *[]Event) error {
	f, e := ioutil.ReadFile(file)
	if e != nil {
		return e
	}
	yaml.Unmarshal([]byte(f), data)
	return e
}

func signalDispatcher() {
	ch := make(chan os.Signal, 1)
	defer close(ch)
	signal.Notify(ch, os.Interrupt, os.Kill)
	for range ch {
		queue.Stop()
	}
}

var queue *Queue
var actions ActionsMap

// Will make retry in sec
var timeout = [...]int{1, 5, 15, 30, 60, 180, 3600, 86400}
var port string = "8080"
var configDir string = "./config"
var queueFile string = "/tmp/worker.state"
var stop chan bool
var queueSize int = 10000

func main() {
	log.Println("[INFO] Elastica Worker!")

	// Load configuration
	actions = NewActionsMap(configDir)

    queue = NewQueue(queueSize, queueFile)

	// Process events queue
	go queue.Process()

	// Catch SIGINT and SIGTERM
	go signalDispatcher()

	// Run HTTP server
	new(HttpServer).Run(port)
}
