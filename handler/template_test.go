package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"

	"obliviate/config"
)

func TestProcessTemplate_ReturnsValidJSON(t *testing.T) {
	variables, err := os.ReadFile("../variables.json")
	if err != nil {
		t.Fatal(err)
	}

	for _, prodEnv := range []bool{false, true} {
		conf := &config.Configuration{
			ProdEnv:            prodEnv,
			EmbededStaticFiles: fstest.MapFS{"variables.json": {Data: variables}},
		}
		handler := ProcessTemplate(conf, "publicKey")

		for _, lang := range []string{"en", "pl", "de", "fr", "tr", "ja", "zh-CN", "zh-TW", "hi", "da"} {
			req := httptest.NewRequest(http.MethodGet, "/variables", nil)
			req.Header.Set("Accept-Language", lang)
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			assert.Equal(t, http.StatusOK, rr.Code, "lang %s", lang)
			var data map[string]any
			if assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &data), "lang %s: invalid JSON", lang) {
				assert.Equal(t, "publicKey", data["PublicKey"], "lang %s", lang)
				assert.NotEmpty(t, data["lang"], "lang %s", lang)
				assert.Contains(t, data["info"], `<a href="https://github.com/blunext/obliviate"`, "lang %s", lang)
				assert.Contains(t, data["info"], `<a href="mailto:info@securenote.io"`, "lang %s", lang)
			}
		}
	}
}

// Run with -race: the template and the cached translation map are shared between requests and must not be modified.
func TestProcessTemplate_ConcurrentRequests(t *testing.T) {
	variables, err := os.ReadFile("../variables.json")
	if err != nil {
		t.Fatal(err)
	}

	for _, prodEnv := range []bool{false, true} {
		conf := &config.Configuration{
			ProdEnv:            prodEnv,
			EmbededStaticFiles: fstest.MapFS{"variables.json": {Data: variables}},
		}
		handler := ProcessTemplate(conf, "publicKey")

		var wg sync.WaitGroup
		for range 50 {
			for _, lang := range []string{"pl", "en", "de"} {
				wg.Go(func() {
					req := httptest.NewRequest(http.MethodGet, "/variables", nil)
					req.Header.Set("Accept-Language", lang)
					rr := httptest.NewRecorder()
					handler.ServeHTTP(rr, req)
					assert.Equal(t, http.StatusOK, rr.Code)
				})
			}
		}
		wg.Wait()
	}
}
