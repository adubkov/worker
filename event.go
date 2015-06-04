package main

import (
    "crypto/md5"
    "os"
    "fmt"
    "log"
    "io/ioutil"
    "gopkg.in/yaml.v2"
    "crypto/rand"
    "path/filepath"
    "regexp"
    "time"
)


// Action for Functions.
type Action struct {
    Type string `json:"type" yaml:"type"`
    Cmd  string `json:"cmd" yaml:"cmd"`
}

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


type Functions map[string][]Action

func selectFunctions(action string) ([]Action, bool) {
    result, ok := functions[action]
    return result, ok
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
    PrintStruct(result)
    return result
}
