package handler

import (
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"unicode/utf8"

	"obliviate/logs"

	"obliviate/config"
)

func StaticFiles(config *config.Configuration, useEmbedFS bool) http.HandlerFunc {
	fs := http.FileServer(getStaticsFS(config.EmbededStaticFiles, useEmbedFS, config.StaticFilesLocation))

	return func(w http.ResponseWriter, r *http.Request) {
		// a path that is not valid UTF-8 cannot name a file, so it is a 404 and not a 500 from http.FileServer
		if !utf8.ValidString(r.URL.Path) {
			http.NotFound(w, r)
			return
		}
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
