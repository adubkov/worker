package main

import (
	"fmt"
	"log"
	"sync"
	"time"
)

// Queue object keep and process events.
type Queue struct {
	queue chan Event
	stop  chan bool
	wg    sync.WaitGroup
	file  string
}

// Make new queue and init it.
func NewQueue(size int, file string) *Queue {
	q := new(Queue)
	q.queue = make(chan Event, size)
	q.stop = make(chan bool, 1)
	q.file = file
	q.load()
	return q
}

// Put event to queue
func (q *Queue) Put(e Event) {
	q.queue <- e
}

// Get event from queue
func (q *Queue) Get() Event {
	return <-q.queue
}

// Stop queue processing
func (q *Queue) Stop() {
	q.stop <- true
}

// Main queue loop, it receive events from channel and process them.
func (q *Queue) Process() {
	defer close(q.queue)
	defer close(q.stop)
	log.Printf("[INFO] Runing queue processing.")
processQueue:
	for {
		select {
		case <-time.After(200 * time.Millisecond):
			select {
			case event := <-q.queue:
				q.wg.Add(1)
				go q.processEvent(event)
			// If receive `true` from stop channel, we should stop processing.
			case <-q.stop:
				log.Printf("[INFO] Stopping queue processing...")
				break processQueue
			}
		}
	}
	// When processing stopped, we should save unprocessed events.
	q.save()
	log.Fatal("[INFO] Shutting down...")
}

// Process event
func (q *Queue) processEvent(event Event) {
	var ok bool = false
	var out string
	defer q.wg.Done()

	// Check if it's time to process event or put it back to queue and return.
	if !q.processRetry(event) {
		return
	}

	// Call apropriate action function.
	switch event.Action.Type {
	case "exec":
		out, ok = actionExec(event)
	case "curl":
		out, ok = actionCurl(event)
	default:
		out = fmt.Sprintf("Can't handle action type: %s", event.Action.Type)
	}

	log.Printf("[INFO] Process event %s try %d: %s", event.Id, event.Attempt, out)

	// If action is not executed successful,
	if !ok {
		// Check if we should back event to queue and wait or drop it.
		if event.Attempt < len(timeout) {
			event.Attempt += 1
			event.Expiration = int64(timeout[event.Attempt-1]) * 1e+9
			q.Put(event)
		} else {
			log.Printf("[INFO] Event %s was dropped: %+v", event.Id, event.Param)
		}
	}
}

// Return true if it's time to try process event, else return false.
func (q *Queue) processRetry(event Event) bool {
	if event.Expiration > 0 {
		event.Expiration -= 200 * 1e+6
		// put it back to queue.
		q.Put(event)
		return false
	}
	return true
}

// Save unprocessed events in file.
func (q *Queue) save() {
	// Make temp variable for events.
	var events []Event
	log.Printf("[INFO] Waiting untill events processing will done")
	q.wg.Wait()
	for {
		select {
		case event := <-q.queue:
			// Fill temp variable with events from queue.
			events = append(events, event)
		default:
			log.Printf("[INFO] Saving events to %s", q.file)
			if nil != saveToFile(q.file, &events) {
				log.Printf("[ERROR] Can not load %s", q.file)
			} else {
				log.Printf("[INFO] Saved %d events", len(events))
			}
			return
		}
	}
}

// Load saved events from file.
func (q *Queue) load() {
	events := []Event{}
	e := loadFromFile(q.file, &events)
	if e != nil {
		log.Printf("[WARN] %s", e)
		return
	}
	// Put events into queue.
	for _, event := range events {
		q.Put(event)
	}
	log.Printf("[INFO] Loaded %d events", len(events))
}
