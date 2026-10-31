package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"

	"obliviate/config"
)

func TestStaticFiles_InvalidPath(t *testing.T) {
	conf := &config.Configuration{
		EmbededStaticFiles:  fstest.MapFS{"build/index.html": {Data: []byte("page")}},
		StaticFilesLocation: "build",
	}
	handler := StaticFiles(conf, true)

	for path, status := range map[string]int{
		"/":          http.StatusOK,
		"/%C0":       http.StatusNotFound,
		"/foo%C0bar": http.StatusNotFound,
		"/%C5%82":    http.StatusNotFound,
	} {
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, path, nil))
		assert.Equal(t, status, rr.Code, "path %s", path)
	}
}
