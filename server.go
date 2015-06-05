package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

// HttpServer object.
type HttpServer struct{}

// Run listener on port.
func (s *HttpServer) Run(port string) {
	log.Printf("[INFO] Runing http server on port: %s", port)
	http.HandleFunc("/", s.rootHandler)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

// Handle request to root ('/').
func (s *HttpServer) rootHandler(res http.ResponseWriter, req *http.Request) {
	statusCode := 400
	result := "FAIL"
	req.ParseForm()
	_, ok := req.Form["id"]
	if ok && len(req.Form["id"][0]) > 0 {
		id := req.Form["id"][0]
		delete(req.Form, "id")

		for _, v := range req.Form {
			for _, param := range v {
				// Get action if it's exist.
				action, ok := actions.Get(param)
				if !ok {
					continue
				}
				// Add event to queue.
				for _, a := range action {
					go s.queueEvent(a, id)
				}
			}
		}
		// Respond as OK if 'id' is exist and not nil.
		statusCode = 200
		result = "OK"
	}
	res.WriteHeader(statusCode)
	fmt.Fprintf(res, result)
}

// Make event and send in to Queue.
func (s *HttpServer) queueEvent(action Action, id string) {
	// Replace $1 in command pattern with 'id'.
	action.Cmd = strings.Replace(action.Cmd, "$1", id, -1)
	// Make new event.
	event := NewEvent(id, action)
	// Put event into the queue.
	log.Printf("[INFO] Add event: %s", event.Id)
	queue.Put(event)
}
