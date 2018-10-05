package main

import (
  "bytes"
	"net/http"
  log "github.com/Sirupsen/logrus"
)

func post(data []byte) {

  // log.Info(string(data))
  r := bytes.NewReader(data)
  response, err := http.Post("http://localhost:8080/v1/data", "application/json", r)
  if err != nil {
    log.Warn(err.Error())
    return
  }

  buf := new(bytes.Buffer)
  buf.ReadFrom(response.Body)
  body := buf.String()

  log.Info("Data posted: ", response.Status, "; ", body)
}
