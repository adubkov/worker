package main

import (
	_ "expvar"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
    "sync"
	"time"
)

func processQueue(queue chan Event, stop chan bool) {
	log.Printf("[INFO] Runing queue processing.")
processQueue:
	for {
		select {
		case event := <-queue:
            wg.Add(1)
			go processEvent(event, queue)
		case <-stop:
			log.Printf("[INFO] Stopping queue processing.")
			break processQueue
		}
	}
	// Dump current queue here.
	saveState(queue)
	log.Fatal("[INFO] Shutting down...")
}

func processRetry(event Event) bool {
	if event.Attempt > 1 {
		retryTime := int64(timeout[event.Attempt-1])
		timeDelta := time.Now().Unix() - event.Timestamp
		if timeDelta <= retryTime {
			queue <- event
			return false
		}
	}
	return true
}

func processEvent(event Event, queue chan Event) {
	var ok bool = false
	var out string
    defer wg.Done()

	if !processRetry(event) {
		return
	}

	switch event.Action.Type {
	case "exec":
		out, ok = actionExec(event)
	case "curl":
		out, ok = actionCurl(event)
	default:
		out = fmt.Sprintf("Can't handle action type: %s", event.Action.Type)
	}

	log.Printf("[INFO] Process event %s try %d: %s", event.Id, event.Attempt, out)

	// If action is not executed successful, return it to queue.
	if !ok {
		if event.Attempt < len(timeout) {
			event.Attempt += 1
			event.Timestamp = time.Now().Unix()
			queue <- event
		} else {
			log.Printf("[INFO] Event %s was dropped: %+v", event.Id, event.Param)
		}
	}
	return
}

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

func saveState(queue chan Event) {
	var events []Event
	log.Printf("[INFO] Waiting untill events processing will done")
    wg.Wait()
	for {
		select {
		case event := <-queue:
			events = append(events, event)
		default:
			log.Printf("[INFO] Saving events to %s", stateFile)
			data, _ := yaml.Marshal(&events)
			saveToFile(stateFile, string(data))
			log.Printf("[INFO] %d events saved", len(events))
			return
		}
	}
}

func signalDispatcher(queue chan Event, stop chan bool) {
	ch := make(chan os.Signal, 1)
	defer close(ch)
	signal.Notify(ch, os.Interrupt, os.Kill)
	for range ch {
		stop <- true
	}
}

var queue chan Event
var actions ActionsMap

// Will make retry in sec
var timeout = [...]int{1, 5, 15, 30, 60, 180, 3600, 86400}
var port string = "8080"
var configDir string = "./config"
var stateFile string = "/tmp/worker.state"
var stop chan bool
var wg sync.WaitGroup

func main() {
	log.Println("[INFO] Elastica Worker!")

	queue = make(chan Event, 10000)
	stop = make(chan bool, 1)
	defer close(queue)
	defer close(stop)

	// Load configuration
    actions = NewActionsMap(configDir)

	// Load saved state
	loadedEvents := []Event{}
	loadFromFile(stateFile, &loadedEvents)
	for _, event := range loadedEvents {
		queue <- event
	}

	// Process events queue
	go processQueue(queue, stop)

	// Catch SIGINT and SIGTERM
	go signalDispatcher(queue, stop)

	// Run HTTP server
    new(HttpServer).Run(port)
}
