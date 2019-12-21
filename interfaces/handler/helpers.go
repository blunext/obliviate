package handler

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
	"net/http"
)

func setStatusAndHeader(w http.ResponseWriter, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
}

func jsonStruct(s interface{}) string {
	j, err := json.Marshal(s)
	if err != nil {
		logrus.Errorf("cannot marshal %v, err: %v", s, err)
	}
	return string(j)
}

func finishRequestWithErr(w http.ResponseWriter, msg string, status int) {
	logrus.Error(msg)
	setStatusAndHeader(w, status)
	w.Write([]byte(""))
}

func finishRequestWithWarn(w http.ResponseWriter, msg string, status int) {
	logrus.Warn(msg)
	setStatusAndHeader(w, status)
	w.Write([]byte(""))
}
