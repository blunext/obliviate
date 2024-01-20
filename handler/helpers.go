package handler

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"obliviate/logs"
)

func setStatusAndHeader(w http.ResponseWriter, status int, prodEnv bool) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
}

func jsonFromStruct(ctx context.Context, s interface{}) []byte {
	j, err := json.Marshal(s)
	if err != nil {
		slog.ErrorContext(ctx, "cannot marshal json", logs.Error, err, logs.JSON, s)
	}
	return j
}

func finishRequestWithErr(ctx context.Context, w http.ResponseWriter, msg string, status int, prodEnv bool) {
	slog.ErrorContext(ctx, msg)
	setStatusAndHeader(w, status, prodEnv)
	//nolint:errcheck
	w.Write([]byte(""))
}

func finishRequestWithWarn(ctx context.Context, w http.ResponseWriter, msg string, status int, prodEnv bool) {
	slog.WarnContext(ctx, msg)
	setStatusAndHeader(w, status, prodEnv)
	//nolint:errcheck
	w.Write([]byte(""))
}
