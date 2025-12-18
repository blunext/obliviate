package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"text/template"

	"obliviate/logs"

	"obliviate/app"
	"obliviate/config"
	"obliviate/handler/webModels"
	"obliviate/i18n"
)

const (
	jsonErrMsg       = "input json error"
	emptyBody        = "empty body post, no json expected"
	maxSaveBodySize  = 256 * 1024 * 6 // ~1.5MB: accounts for base64 encoding (~33% overhead) + JSON structure
	maxSmallBodySize = 1024           // for Read/Delete endpoints (hash + public key)
)

func ProcessTemplate(config *config.Configuration, publicKey string) http.HandlerFunc {

	var t *template.Template
	if config.ProdEnv {
		t = template.Must(template.New("variables.json").ParseFS(config.EmbededStaticFiles, "variables.json"))
	}

	translation := i18n.NewTranslation()

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		slog.InfoContext(ctx, "ProcessTemplate Handler")

		if !config.ProdEnv {
			t, _ = template.New("variables.json").ParseFS(config.EmbededStaticFiles, "variables.json")
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("Vary", "Accept-Language")

		data := translation.GetTranslation(ctx, r.Header.Get("Accept-Language"))
		data["PublicKey"] = publicKey
		err := t.Execute(w, data)
		if err != nil {
			slog.ErrorContext(ctx, "Could not execute the template", logs.Error, err, logs.TemplateData, data)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}
}

func Save(app *app.App) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		slog.InfoContext(ctx, "Save handler...")

		r.Body = http.MaxBytesReader(w, r.Body, maxSaveBodySize)
		defer r.Body.Close()
		if r.Body == nil {
			finishRequestWithErr(ctx, w, emptyBody, http.StatusBadRequest, app.Config.ProdEnv)
			return
		}

		data := webModels.SaveRequest{}
		err := json.NewDecoder(r.Body).Decode(&data)
		switch {
		case err != nil:
			finishRequestWithErr(ctx, w, jsonErrMsg, http.StatusBadRequest, app.Config.ProdEnv)
		case len(data.Message) == 0 || len(data.Message) > 256*1024*4:
			finishRequestWithErr(ctx, w, fmt.Sprintf("Message len is wrong = %d", len(data.Message)), http.StatusBadRequest, app.Config.ProdEnv)
		case len(data.TransmissionNonce) == 0:
			finishRequestWithErr(ctx, w, "TransmissionNonce is empty", http.StatusBadRequest, app.Config.ProdEnv)
		case len(data.Hash) == 0:
			finishRequestWithErr(ctx, w, "Hash is empty", http.StatusBadRequest, app.Config.ProdEnv)
		case len(data.TransmissionNonce) != 24:
			finishRequestWithErr(ctx, w, "TransmissionNonce length is wrong !=24", http.StatusBadRequest, app.Config.ProdEnv)
		case len(data.PublicKey) != 32:
			finishRequestWithErr(ctx, w, "PublicKey length is wrong !=32", http.StatusBadRequest, app.Config.ProdEnv)
		default:
			ctx := context.WithValue(ctx, config.AcceptLanguage, r.Header.Get("Accept-Language"))
			ctx = context.WithValue(ctx, config.CountryCode, r.Header.Get("CF-IPCountry"))
			err = app.ProcessSave(ctx, data)
			if err != nil {
				finishRequestWithErr(ctx, w, fmt.Sprintf("Cannot process input message, err: %v", err), http.StatusBadRequest, app.Config.ProdEnv)
				return
			}
			setStatusAndHeader(w, http.StatusOK, app.Config.ProdEnv)
			//nolint:errcheck
			w.Write([]byte("[]"))
		}
	}
}

func Read(app *app.App) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		slog.InfoContext(ctx, "Read handler")

		r.Body = http.MaxBytesReader(w, r.Body, maxSmallBodySize)
		defer r.Body.Close()
		if r.Body == nil {
			finishRequestWithErr(ctx, w, emptyBody, http.StatusBadRequest, app.Config.ProdEnv)
			return
		}

		data := webModels.ReadRequest{}
		err := json.NewDecoder(r.Body).Decode(&data)
		switch {
		case err != nil:
			finishRequestWithErr(ctx, w, jsonErrMsg, http.StatusBadRequest, app.Config.ProdEnv)
		case len(data.Hash) == 0:
			finishRequestWithErr(ctx, w, "Hash not found", http.StatusBadRequest, app.Config.ProdEnv)
		case len(data.PublicKey) == 0:
			finishRequestWithErr(ctx, w, "PublicKey not found", http.StatusBadRequest, app.Config.ProdEnv)
		case len(data.PublicKey) != 32:
			finishRequestWithErr(ctx, w, "PublicKey length is wrong !=32", http.StatusBadRequest, app.Config.ProdEnv)
		default:
			encrypted, costFactor, err := app.ProcessRead(ctx, data)
			if err != nil {
				finishRequestWithErr(ctx, w, fmt.Sprintf("Cannot process read message, err: %v", err), http.StatusBadRequest, app.Config.ProdEnv)
				return
			}
			if encrypted == nil {
				// not found
				finishRequestWithWarn(ctx, w, "Message not found", http.StatusNotFound, app.Config.ProdEnv)
				return
			}

			message := webModels.ReadResponse{Message: encrypted, CostFactor: costFactor}

			setStatusAndHeader(w, http.StatusOK, app.Config.ProdEnv)
			//nolint:errcheck
			w.Write(jsonFromStruct(ctx, message))
		}
	}
}

func Delete(app *app.App) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		slog.InfoContext(ctx, "Delete Handler")

		r.Body = http.MaxBytesReader(w, r.Body, maxSmallBodySize)
		defer r.Body.Close()
		if r.Body == nil {
			finishRequestWithErr(ctx, w, emptyBody, http.StatusBadRequest, app.Config.ProdEnv)
			return
		}

		data := webModels.DeleteRequest{}
		err := json.NewDecoder(r.Body).Decode(&data)
		if err != nil {
			finishRequestWithErr(ctx, w, jsonErrMsg, http.StatusBadRequest, app.Config.ProdEnv)
			return
		}
		if len(data.Hash) == 0 {
			finishRequestWithErr(ctx, w, "Hash is empty", http.StatusBadRequest, app.Config.ProdEnv)
			return
		}

		app.ProcessDelete(r.Context(), data.Hash)

		setStatusAndHeader(w, http.StatusOK, app.Config.ProdEnv)
		//nolint:errcheck
		w.Write([]byte("[]"))
	}
}

func Expired(app *app.App) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		slog.InfoContext(ctx, "Expired handler")

		if err := app.ProcessDeleteExpired(ctx); err != nil {
			slog.ErrorContext(ctx, "delete expired error", logs.Error, err.Error())
			setStatusAndHeader(w, http.StatusInternalServerError, app.Config.ProdEnv)
		} else {
			setStatusAndHeader(w, http.StatusOK, app.Config.ProdEnv)
		}
		//nolint:errcheck
		w.Write([]byte("[]"))
	}
}
