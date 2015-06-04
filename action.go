package main

import (
    "os"
    "log"
    "io/ioutil"
    "gopkg.in/yaml.v2"
    "path/filepath"
    "regexp"
)

// Action for Functions.
type Action struct {
    Type string `json:"type" yaml:"type"`
    Cmd  string `json:"cmd" yaml:"cmd"`
}

type Actions map[string][]Action
type ActionsMap struct{
    list Actions
}

func NewActionsMap(path string) ActionsMap {
    am := ActionsMap{}
    am.list = make(map[string][]Action)
    am.loadActions(path)
    return am
}

func (am *ActionsMap) Get(action string) ([]Action, bool) {
    result, ok := am.list[action]
    return result, ok
}

// Load all functions from directory and combine them to []Functions{}.
func (am *ActionsMap) loadActions(path string) {
    result := Actions{}

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
        t, e := am.loadFromFile(file_name, Actions{})
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
    am.list = result
}

// Load from YAML file. Actually work only with Functions data type.
func (am *ActionsMap) loadFromFile(file string, data Actions) (Actions, error) {
    f, e := ioutil.ReadFile(file)
    if e != nil {
        log.Fatal("[CRITICAL] File read error: %v", e)
    }
    e = yaml.Unmarshal(f, &data)
    return data, e
}