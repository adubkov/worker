package main

import (
	"bytes"
	"encoding/json"
    "gopkg.in/yaml.v2"
    "os"
    "io/ioutil"
	"log"
)

func saveToFile(file string, data interface{}) error {
    f, e := os.Create(file)
    defer f.Close()
    if e != nil {
        return e
    }
    strData, e := yaml.Marshal(data)
    if e != nil {
        return e
    }
    f.WriteString(string(strData))
    return e
}

func loadFromFile(file string, data interface{}) error {
    f, e := ioutil.ReadFile(file)
    if e != nil {
        return e
    }
    yaml.Unmarshal([]byte(f), data)
    return e
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
