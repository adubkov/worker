package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

type HttpServer struct{}

func (s *HttpServer) Run(port string) {
	log.Printf("[INFO] Runing http server on port: %s", port)
	http.HandleFunc("/", s.rootHandler)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func (s *HttpServer) rootHandler(res http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	log.Printf("[INFO] Parsed request: %s", req.URL)
	_, ok := req.Form["id"]
	if ok && len(req.Form["id"][0]) > 0 {
		id := req.Form["id"][0]
		delete(req.Form, "id")

		for _, v := range req.Form {
			for _, action := range v {
				function, ok := actions.Get(action)
				if !ok {
					continue
				}
				// Add event to queue.
				for _, a := range function {
					go s.queueEvent(a, id)
				}
			}
		}
		// respond as OK if 'id' is exist and not nil
		fmt.Fprintf(res, "OK")
	} else {
		fmt.Fprintf(res, "FAIL")
	}
}

func (s *HttpServer) queueEvent(action Action, id string) {
	// Replace $1 in command pattern with id.
	action.Cmd = strings.Replace(action.Cmd, "$1", id, -1)
	event := NewEvent(id, action)
	log.Printf("[INFO] Add event: %s", event.Id)
	queue <- event
}
