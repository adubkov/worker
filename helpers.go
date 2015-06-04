package main

import (
	"bytes"
	"encoding/json"
	"log"
)

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
