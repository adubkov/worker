package main

import (
	"log"
	"os"
	"path/filepath"
	"regexp"
)

// Describe type of action and contain command to execute.
type Action struct {
	Type string `json:"type" yaml:"type"`
	Cmd  string `json:"cmd" yaml:"cmd"`
}

// Set of actions.
type Actions map[string][]Action

// Object ActionsMap provice interface to get actions by name.
type ActionsMap struct {
	list Actions
}

// Make new ActionsMap object and initialize it.
// It loads actions list from the files in `path`.
func NewActionsMap(path string) ActionsMap {
	am := ActionsMap{}
	am.list = make(map[string][]Action)
	am.loadActions(path)
	return am
}

// Return action object by name.
func (am *ActionsMap) Get(action string) ([]Action, bool) {
	result, ok := am.list[action]
	return result, ok
}

// Search files in path, by regexp patthern.
func (am *ActionsMap) search(path, pattern string) []string {
	var files []string
	// Walk thought all files in path.
	filepath.Walk(path, func(p string, f os.FileInfo, _ error) error {
		if !f.IsDir() {
			isYamlFile, _ := regexp.MatchString(pattern, f.Name())
			if isYamlFile {
				files = append(files, f.Name())
			}
		}
		return nil
	})
	log.Printf("[INFO] Found %d files to load from '%s'", len(files), path)
	return files
}

// Load all actions files and combine them.
func (am *ActionsMap) loadActions(path string) {
	result := Actions{}

	// Find all .yaml files in path.
	files := am.search(path, ".yaml$")

	// Load all yaml files in path.
	for _, file := range files {
		file_name := path + "/" + file
		a := Actions{}
		if nil != loadFromFile(file_name, &a) {
			log.Fatalf("[ERROR] Can not load %s", file_name)
		}
		log.Printf("[INFO] Loading actions from: %s", file_name)

		// Append actions to result.
		for k, v := range a {
			result[k] = v
			log.Printf("[INFO] Register action: %s", k)
		}
	}

	log.Printf("[INFO] Registered %d actions", len(result))
	am.list = result
}
