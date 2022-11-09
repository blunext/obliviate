package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"text/template"

	"github.com/sirupsen/logrus"

	"obliviate/app"
	"obliviate/config"
	"obliviate/handler/webModels"
	"obliviate/i18n"
)

const (
	jsonErrMsg = "input json error"
	emptyBody  = "empty body post, no json expected"
)

func ProcessTemplate(config *config.Configuration, publicKey string) http.HandlerFunc {

	var t *template.Template
	if config.ProdEnv {
		t = template.Must(template.New("variables.json").ParseFS(config.EmbededStaticFiles, "variables.json"))
	}

	translation := i18n.NewTranslation()

	return func(w http.ResponseWriter, r *http.Request) {
		logrus.Trace("ProcessTemplate Handler")

		if !config.ProdEnv {
			t, _ = template.New("variables.json").ParseFS(config.EmbededStaticFiles, "variables.json")
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("Cache-Control", "no-cache, no-store")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		data := translation.GetTranslation(r.Header.Get("Accept-Language"))
		data["PublicKey"] = publicKey
		t.Execute(w, data)
	}
}

func Save(app *app.App) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		logrus.Trace("Save handler...")

		defer r.Body.Close()
		if r.Body == nil {
			finishRequestWithErr(w, emptyBody, http.StatusBadRequest)
			return
		}

		data := webModels.SaveRequest{}
		err := json.NewDecoder(r.Body).Decode(&data)
		switch {
		case err != nil:
			finishRequestWithErr(w, jsonErrMsg, http.StatusBadRequest)
		case len(data.Message) == 0 || len(data.Message) > 256*1024*4:
			finishRequestWithErr(w, fmt.Sprintf("Message len is wrong = %d", len(data.Message)), http.StatusBadRequest)
		case len(data.TransmissionNonce) == 0:
			finishRequestWithErr(w, "TransmissionNonce is empty", http.StatusBadRequest)
		case len(data.Hash) == 0:
			finishRequestWithErr(w, "Hash is empty", http.StatusBadRequest)
		case len(data.TransmissionNonce) != 24:
			finishRequestWithErr(w, "TransmissionNonce length is wrong !=24", http.StatusBadRequest)
		case len(data.PublicKey) != 32:
			finishRequestWithErr(w, "PublicKey length is wrong !=24", http.StatusBadRequest)
		default:
			err = app.ProcessSave(r.Context(), data, r.Header.Get("CF-IPCountry"))
			if err != nil {
				finishRequestWithErr(w, fmt.Sprintf("Cannot process input message, err: %v", err), http.StatusBadRequest)
				return
			}
			setStatusAndHeader(w, http.StatusOK)
			w.Write([]byte("[]"))
		}
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

		data := webModels.ReadRequest{}
		err := json.NewDecoder(r.Body).Decode(&data)
		switch {
		case err != nil:
			finishRequestWithErr(w, jsonErrMsg, http.StatusBadRequest)
		case len(data.Hash) == 0:
			finishRequestWithErr(w, "Hash not found", http.StatusBadRequest)
		case len(data.PublicKey) == 0:
			finishRequestWithErr(w, "PublicKey not found", http.StatusBadRequest)
		case len(data.PublicKey) != 32:
			finishRequestWithErr(w, "PublicKey length is wrong !=32", http.StatusBadRequest)
		default:
			encrypted, costFactor, err := app.ProcessRead(r.Context(), data)
			if err != nil {
				finishRequestWithErr(w, fmt.Sprintf("Cannot process read message, err: %v", err), http.StatusBadRequest)
				return
			}
			if encrypted == nil {
				// not found
				finishRequestWithWarn(w, "Message not found", http.StatusNotFound)
				return
			}

			message := webModels.ReadResponse{Message: encrypted, CostFactor: costFactor}

			setStatusAndHeader(w, http.StatusOK)
			w.Write(jsonFromStruct(message))
		}
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

		data := webModels.DeleteRequest{}
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
