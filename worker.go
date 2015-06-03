package main

import (
	"bytes"
	"crypto/md5"
	"crypto/rand"
	"encoding/json"
	_ "expvar"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// Action for Functions.
type Action struct {
	Type string `json:"type" yaml:"type"`
	Cmd  string `json:"cmd" yaml:"cmd"`
}

type Functions map[string][]Action

type Event struct {
	Id        string
	Timestamp int64
	Param     string
	Action    Action
	Attempt   int
}

// Load YAML from file. Actually work only with Functions data type.
func loadYAMLFromFile(file string, data Functions) (Functions, error) {
	f, e := ioutil.ReadFile(file)
	if e != nil {
		log.Fatal("[CRITICAL] File read error: %v", e)
	}
	e = yaml.Unmarshal(f, &data)
	return data, e
}

// Load all functions from directory and combine them to []Functions{}.
func loadFunctions(path string) Functions {
	result := Functions{}

	// Find all .yaml files in path
	var files []string
	filepath.Walk(path, func(p string, f os.FileInfo, _ error) error {
		if !f.IsDir() {
			isYamlFile, _ := regexp.MatchString(".yaml$", f.Name())
			if isYamlFile {
				files = append(files, f.Name())
			}
		}
		return nil
	})

	log.Printf("[INFO] Found %d files to load from '%s'", len(files), path)

	// Load all yaml files in dir
	for _, file := range files {
		file_name := path + "/" + file
		t, e := loadYAMLFromFile(file_name, Functions{})
		if e != nil {
			log.Printf("[FAILED] Loading %s", file_name)
			os.Exit(1)
		}
		log.Printf("[INFO] Loaded %s", file_name)

        // Merge
		for k, v := range t {
			result[k] = v
		}
	}

	log.Printf("[INFO] %d Functions loaded", len(result))
	return result
}

// Pretty print structure as YAML.
func PrintStruct(v interface{}) {
	var out bytes.Buffer

	data, e := json.Marshal(v)
	if e != nil {
		log.Fatal("[CRITICAL] JSON Marshal error: %v", e)
	}

	if e = json.Indent(&out, data, "", "  "); e != nil {
		log.Fatal("[CRITICAL] JSON Indent error: %v", e)
	}

	log.Printf("%s\n", out.String())
}

func getId() string {
	random := make([]byte, 1024)
	rand.Read(random)
	hash := fmt.Sprintf("%x", md5.Sum(random))
	return hash
}

func selectFunctions(action string) ([]Action, bool) {
	result, ok := functions[action]
	return result, ok
}

func queueEvent(action Action, id string) {
	// Replace $1 in command pattern with id.
	action.Cmd = strings.Replace(action.Cmd, "$1", id, -1)
	timestamp := time.Now().Unix()
	event := Event{Id: getId(), Timestamp: timestamp, Param: id, Action: action, Attempt: 1}
	log.Printf("[INFO] Add event: %s", event.Id)
	queue <- event
}

func httpHandler(res http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	log.Printf("[INFO] Parsed request: %s", req.URL)
	_, ok := req.Form["id"]
	if ok && len(req.Form["id"][0]) > 0 {
		id := req.Form["id"][0]
		delete(req.Form, "id")

		for _, v := range req.Form {
			for _, action := range v {
				function, ok := selectFunctions(action)
				if !ok {
					continue
				}
				// Add event to queue.
				for _, a := range function {
					go queueEvent(a, id)
				}
			}
		}
		// respond as OK if 'id' is exist and not nil
		fmt.Fprintf(res, "OK")
	} else {
		fmt.Fprintf(res, "FAIL")
	}
}

func httpServerRun(port string) {
	log.Printf("[INFO] Runing http server on port: %s", port)
	http.HandleFunc("/", httpHandler)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func processQueue(queue chan Event, stop chan bool) {
	log.Printf("[INFO] Runing queue processing.")
processQueue:
	for {
		select {
		case event := <-queue:
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

func actionExec(event Event) (string, bool) {
	// Make command structure
	cmd := exec.Command("sh", "-c", event.Action.Cmd)
	// Execute process
	out, e := cmd.Output()
	if e != nil {
		return "FAIL", false
	}
	return string(out), true
}

func actionCurl(event Event) (string, bool) {
    // Set timeout for connection
	client := http.Client{
		Timeout: time.Duration(10 * time.Second),
	}
	// Make http request.
	res, e := client.Get(event.Action.Cmd)
	if e != nil {
		return "FAIL", false
	}
	// If respond success, we are ok.
	if res.StatusCode != 200 {
		return "FAIL", false
	}
	defer res.Body.Close()
	out, _ := ioutil.ReadAll(res.Body)
	return string(out), true
}

func processEvent(event Event, queue chan Event) {
	var ok bool = false
	var out string

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
	time.Sleep(time.Duration(5 * time.Second))
	for {
		select {
		case event := <-queue:
			events = append(events, event)
		default:
			log.Printf("[INFO] Saving events to %s", stateFile)
			data, _ := yaml.Marshal(&events)
			saveToFile(stateFile, string(data))
			log.Printf("[INFO] Saved %d events", len(events))
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
var functions Functions

// Will make retry in sec
var timeout = [...]int{1, 5, 15, 30, 60, 180, 3600, 86400}
var port string = "8080"
var configDir string = "./config"
var stateFile string = "/tmp/worker.state"
var stop chan bool

func main() {
	log.Println("[INFO] Elastica Worker!")

	queue = make(chan Event, 10000)
	stop = make(chan bool, 1)
	defer close(queue)
	defer close(stop)

	// Load configuration
	functions = loadFunctions(configDir)

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
	httpServerRun(port)
}
