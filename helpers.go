package main

import (
	"bytes"
	"encoding/json"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
)

// Save data to file in yaml format.
func saveToFile(file string, data interface{}) error {
	strData, e := yaml.Marshal(data)
	if e != nil {
		return e
	}
	e = ioutil.WriteFile(file, strData, 0644)
	if e != nil {
		return e
	}
	return nil
}

// Load data from file in yaml format.
// Object structure should match yaml structure.
func loadFromFile(file string, data interface{}) error {
	f, e := ioutil.ReadFile(file)
	if e != nil {
		return e
	}
	yaml.Unmarshal([]byte(f), data)
	return nil
}

// Pretty print structure as YAML.
func prettyPrint(v interface{}) {
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
