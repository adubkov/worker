package main

import (
    "io/ioutil"
    "net/http"
    "os/exec"
    "time"
)

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
        return "http.Get error", false
    }
    // If respond success, we are ok.
    if res.StatusCode != 200 {
        return string(res.StatusCode), false
    }
    defer res.Body.Close()
    out, _ := ioutil.ReadAll(res.Body)
    return string(out), true
}
