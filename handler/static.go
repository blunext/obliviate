package handler

import (
	"io/fs"
	"log/slog"
	"net/http"
	"obliviate/logs"
	"os"

	"obliviate/config"
)

func StaticFiles(config *config.Configuration, useEmbedFS bool) http.HandlerFunc {
	fs := http.FileServer(getStaticsFS(config.EmbededStaticFiles, useEmbedFS, config.StaticFilesLocation))

	return func(w http.ResponseWriter, r *http.Request) {
		fs.ServeHTTP(w, r)
	}
}

func getStaticsFS(static fs.FS, useEmbedFS bool, stripPath string) http.FileSystem {
	if !useEmbedFS {
		slog.Info("using live os files mode")
		return http.FS(os.DirFS(stripPath))
	}

	slog.Info("using embed files mode")
	fsys, err := fs.Sub(static, stripPath)
	if err != nil {
		slog.Error("FS error", logs.Error, err)
	}

	return http.FS(fsys)
}
