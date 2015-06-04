package main

import (
    "fmt"
    "time"
    "log"
    "gopkg.in/yaml.v2"
)

type Queue struct {
    queue chan Event
    stop chan bool
}

func NewQueue(size int) *Queue {
    q := new(Queue)
    q.queue = make(chan Event, size)
    q.stop = make(chan bool, 1)
    return q
}

func (q *Queue) Put(e Event) {
    q.queue <- e
}

func (q *Queue) Get() Event {
    return <- q.queue
}

func (q *Queue) Stop() {
    q.stop <- true
}

func (q *Queue) Process() {
    defer close(q.queue)
    defer close(q.stop)
    log.Printf("[INFO] Runing queue processing.")
processQueue:
    for {
        select {
        case event := <-q.queue:
            wg.Add(1)
            go q.processEvent(event)
        case <- q.stop:
            log.Printf("[INFO] Stopping queue processing.")
            break processQueue
        }
    }
    // Dump current queue here.
    q.save()
    log.Fatal("[INFO] Shutting down...")
}

func (q *Queue) processEvent(event Event) {
    var ok bool = false
    var out string
    defer wg.Done()

    if !q.processRetry(event) {
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
            q.Put(event)
        } else {
            log.Printf("[INFO] Event %s was dropped: %+v", event.Id, event.Param)
        }
    }
    return
}

func (q *Queue) processRetry(event Event) bool {
    if event.Attempt > 1 {
        retryTime := int64(timeout[event.Attempt-1])
        timeDelta := time.Now().Unix() - event.Timestamp
        if timeDelta <= retryTime {
            q.Put(event)
            return false
        }
    }
    return true
}

func (q *Queue) save() {
    var events []Event
    log.Printf("[INFO] Waiting untill events processing will done")
    wg.Wait()
    for {
        select {
        case event := <-q.queue:
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
