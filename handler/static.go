package handler

import (
	"github.com/sirupsen/logrus"
	"io/fs"
	"net/http"
	"obliviate/config"
	"os"
)

func StaticFiles(config *config.Configuration, useEmbedFS bool) http.HandlerFunc {
	fs := http.FileServer(getStaticsFS(config.EmbededStaticFiles, useEmbedFS, config.StaticFilesLocation))

	return func(w http.ResponseWriter, r *http.Request) {
		fs.ServeHTTP(w, r)
	}
}

func getStaticsFS(static fs.FS, useEmbedFS bool, stripPath string) http.FileSystem {
	if !useEmbedFS {
		logrus.Println("using live os files mode")
		return http.FS(os.DirFS(stripPath))
	}

	logrus.Println("using embed files mode")
	fsys, err := fs.Sub(static, stripPath)
	if err != nil {
		logrus.Panicln(err)
	}

	return http.FS(fsys)
}
