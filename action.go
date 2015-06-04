package main

import (
	"log"
	"os"
	"path/filepath"
	"regexp"
)

// Action for Functions.
type Action struct {
	Type string `json:"type" yaml:"type"`
	Cmd  string `json:"cmd" yaml:"cmd"`
}

type Actions map[string][]Action
type ActionsMap struct {
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
		a := Actions{}
		if nil != loadFromFile(file_name, &a) {
			log.Fatalf("[ERROR] Can not load %s", file_name)
		}
		log.Printf("[INFO] Loaded %s", file_name)

		// Merge
		for k, v := range a {
			result[k] = v
		}
	}

	log.Printf("[INFO] %d Functions loaded", len(result))
	PrintStruct(result)
	am.list = result
}
