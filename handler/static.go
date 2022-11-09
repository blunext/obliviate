package handler

import (
	"io/fs"
	"net/http"
	"os"

	"github.com/sirupsen/logrus"

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
		logrus.Trace("using live os files mode")
		return http.FS(os.DirFS(stripPath))
	}

	logrus.Trace("using embed files mode")
	fsys, err := fs.Sub(static, stripPath)
	if err != nil {
		logrus.Panicln(err)
	}

	return http.FS(fsys)
}
