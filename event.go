package main

import (
	"crypto/md5"
	"crypto/rand"
	"fmt"
	"time"
)

// Describe event which should be happen.
// `Param` now mostly used for specify id of host.
type Event struct {
	Id        string
	Timestamp int64
	Param     string 
	Action    Action
	Attempt   int
}

// Make new event and initialize it.
func NewEvent(id string, action Action) Event {
	ev := Event{
		Timestamp: time.Now().Unix(),
		Param:     id,
		Action:    action,
		Attempt:   1}
	ev.Id = ev.NewId()
	return ev
}

// Return random md5 hash.
func (ev *Event) NewId() string {
	random := make([]byte, 1024)
	rand.Read(random)
	hash := fmt.Sprintf("%x", md5.Sum(random))
	return hash
}
