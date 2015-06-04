package main

import (
    "crypto/md5"
    "fmt"
    "crypto/rand"
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
    ev := Event{}
    ev.Id = ev.NewId()
    ev.Timestamp = time.Now().Unix()
    ev.Param = id
    ev.Action = action
    ev.Attempt = 1
    return ev
}

func (ev *Event) NewId() string {
    random := make([]byte, 1024)
    rand.Read(random)
    hash := fmt.Sprintf("%x", md5.Sum(random))
    return hash
}
