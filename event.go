package main

import (
	"crypto/md5"
	"crypto/rand"
	"fmt"
	"time"
)

type Event struct {
	Id        string
	Timestamp int64
	Param     string
	Action    Action
	Attempt   int
}

func NewEvent(id string, action Action) Event {
	ev := Event{
		Timestamp: time.Now().Unix(),
		Param:     id,
		Action:    action,
		Attempt:   1}
	ev.Id = ev.NewId()
	return ev
}

func (ev *Event) NewId() string {
	random := make([]byte, 1024)
	rand.Read(random)
	hash := fmt.Sprintf("%x", md5.Sum(random))
	return hash
}
