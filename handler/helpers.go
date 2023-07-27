package handler

import (
	"encoding/json"
	"net/http"

	"github.com/sirupsen/logrus"
)

func setStatusAndHeader(w http.ResponseWriter, status int, prodEnv bool) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
}

func jsonFromStruct(s interface{}) []byte {
	j, err := json.Marshal(s)
	if err != nil {
		logrus.Errorf("cannot marshal %v, err: %v", s, err)
	}
	return j
}

func finishRequestWithErr(w http.ResponseWriter, msg string, status int, prodEnv bool) {
	logrus.Error(msg)
	setStatusAndHeader(w, status, prodEnv)
	w.Write([]byte(""))
}

func finishRequestWithWarn(w http.ResponseWriter, msg string, status int, prodEnv bool) {
	logrus.Warn(msg)
	setStatusAndHeader(w, status, prodEnv)
	w.Write([]byte(""))
}
