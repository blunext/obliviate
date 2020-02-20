package handler

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"html/template"
	"net/http"
	"obliviate/app"
	"obliviate/config"
	"obliviate/i18n"
)

type SaveRequest struct {
	Message           []byte `json:"message"`
	TransmissionNonce []byte `json:"nonce"`
	Hash              string `json:"hash"`
	PublicKey         []byte `json:"publicKey"`
	Time              int    `json:"time"`
}

type ReadRequest struct {
	Hash      string `json:"hash"`
	PublicKey []byte `json:"publicKey"`
	Password  bool   `json:"password"`
}

type DeleteRequest struct {
	Hash string `json:"hash"`
}

type ReadResponse struct {
	Message []byte `json:"message"`
}

const (
	jsonErrMsg = "input json error"
	emptyBody  = "empty body post, no json expected"
)

func ProcessTemplate(config *config.Configuration, publicKey string) http.HandlerFunc {

	var t *template.Template
	if config.ProdEnv {
		t = template.Must(template.New("template.html").ParseFiles("template.html"))
	}

	translation := i18n.NewTranslation()

	return func(w http.ResponseWriter, r *http.Request) {
		logrus.Trace("ProcessTemplate Handler")

		if !config.ProdEnv {
			t, _ = template.New("template.html").ParseFiles("template.html")
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-cache, no-store")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")

		data := translation.GetTranslation(r.Header.Get("Accept-Language"))
		data["PublicKey"] = publicKey
		t.Execute(w, data)
	}
}

func Save(app *app.App) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		logrus.Trace("Save Handler")

		defer r.Body.Close()
		if r.Body == nil {
			finishRequestWithErr(w, emptyBody, http.StatusBadRequest)
			return
		}

		data := SaveRequest{}
		err := json.NewDecoder(r.Body).Decode(&data)
		if err != nil {
			finishRequestWithErr(w, jsonErrMsg, http.StatusBadRequest)
			return
		}
		if len(data.Message) == 0 || len(data.Message) > 300000 {
			// 282818 is encoded length of 256k of txt
			finishRequestWithErr(w, fmt.Sprintf("Message len is wrong = %d", len(data.Message)), http.StatusBadRequest)
			return
		}
		if len(data.TransmissionNonce) == 0 {
			finishRequestWithErr(w, "TransmissionNonce is empty", http.StatusBadRequest)
			return
		}
		if len(data.Hash) == 0 {
			finishRequestWithErr(w, "Hash is empty", http.StatusBadRequest)
			return
		}

		if len(data.TransmissionNonce) != 24 {
			finishRequestWithErr(w, "TransmissionNonce length is wrong !=24", http.StatusBadRequest)
			return
		}
		if len(data.PublicKey) != 32 {
			finishRequestWithErr(w, "PublicKey length is wrong !=24", http.StatusBadRequest)
			return
		}

		err = app.ProcessSave(r.Context(), data.Message, data.TransmissionNonce, data.Hash, data.PublicKey, data.Time)
		if err != nil {
			finishRequestWithErr(w, fmt.Sprintf("Cannot process input message, err: %v", err), http.StatusBadRequest)
			return
		}

		setStatusAndHeader(w, http.StatusOK)
		w.Write([]byte("[]"))
	}
}

func Read(app *app.App) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		logrus.Trace("Read handler")

		defer r.Body.Close()
		if r.Body == nil {
			finishRequestWithErr(w, emptyBody, http.StatusBadRequest)
			return
		}

		data := ReadRequest{}
		err := json.NewDecoder(r.Body).Decode(&data)
		if err != nil {
			finishRequestWithErr(w, jsonErrMsg, http.StatusBadRequest)
			return
		}
		if len(data.Hash) == 0 {
			finishRequestWithErr(w, "Hash not found", http.StatusBadRequest)
			return
		}
		if len(data.PublicKey) == 0 {
			finishRequestWithErr(w, "PublicKey not found", http.StatusBadRequest)
			return
		}
		if len(data.PublicKey) != 32 {
			finishRequestWithErr(w, "PublicKey length is wrong !=32", http.StatusBadRequest)
			return
		}

		encrypted, err := app.ProcessRead(r.Context(), data.Hash, data.PublicKey, data.Password)
		if err != nil {
			finishRequestWithErr(w, fmt.Sprintf("Cannot process read message, err: %v", err), http.StatusBadRequest)
			return
		}
		if encrypted == nil {
			// not found
			finishRequestWithWarn(w, "Message not found", http.StatusNotFound)
			return
		}

		message := ReadResponse{Message: encrypted}

		setStatusAndHeader(w, http.StatusOK)
		w.Write(jsonFromStruct(message))

	}
}

func Delete(app *app.App) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		logrus.Trace("Delete Handler")

		defer r.Body.Close()
		if r.Body == nil {
			finishRequestWithErr(w, emptyBody, http.StatusBadRequest)
			return
		}

		data := DeleteRequest{}
		err := json.NewDecoder(r.Body).Decode(&data)
		if err != nil {
			finishRequestWithErr(w, jsonErrMsg, http.StatusBadRequest)
			return
		}
		if len(data.Hash) == 0 {
			finishRequestWithErr(w, "Hash is empty", http.StatusBadRequest)
			return
		}

		app.ProcessDelete(r.Context(), data.Hash)

		setStatusAndHeader(w, http.StatusOK)
		w.Write([]byte("[]"))
	}
}

func Expired(app *app.App) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		logrus.Trace("Expired handler")

		if err := app.ProcessDeleteExpired(r.Context()); err != nil {
			logrus.Errorf(err.Error())
			setStatusAndHeader(w, http.StatusInternalServerError)
		} else {
			setStatusAndHeader(w, http.StatusOK)
		}
		w.Write([]byte("[]"))
	}
}
