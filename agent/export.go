package agent

import (
  "bytes"
	"net/http"
  log "github.com/Sirupsen/logrus"
)

func post(data []byte, serverAddress string, token string) {

  // Convert json []byte to Reader
  r := bytes.NewReader(data)

  client := &http.Client{}
  req, err := http.NewRequest("POST", serverAddress + "/v1/data", r)
  req.Header.Add("X-Kubeviz-Token", token)
  req.Header.Add("Content-Type", "application/json")
  response, err := client.Do(req)

  if err != nil {
    log.Warn(err.Error())
    return
  }

  buf := new(bytes.Buffer)
  buf.ReadFrom(response.Body)
  body := buf.String()

  log.Info("Data posted: ", response.Status, "; ", body)
}
