package main

import (
	"context"
	"embed"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/NYTimes/gziphandler"
	_ "golang.org/x/crypto/x509roots/fallback"

	"obliviate/logs"

	"obliviate/app"
	"obliviate/config"
	"obliviate/crypt"
	"obliviate/crypt/rsa"
	"obliviate/handler"
	"obliviate/repository"
	"obliviate/repository/mock"
)

const (
	messageDurationTime = time.Hour * 24 * 7 * 4
)

//go:embed variables.json
//go:embed all:web/build
var static embed.FS

func main() {
	conf := config.Configuration{
		DefaultDurationTime:     messageDurationTime,
		ProdEnv:                 os.Getenv("ENV") == "PROD",
		MasterKey:               os.Getenv("HSM_MASTER_KEY"),
		KmsCredentialFile:       os.Getenv("KMS_CREDENTIAL_FILE"),
		FirestoreCredentialFile: os.Getenv("FIRESTORE_CREDENTIAL_FILE"),
		StaticFilesLocation:     "web/build",
		EmbededStaticFiles:      static,
	}

	var algorithm rsa.EncryptionOnRest
	var db repository.DataBase

	if conf.ProdEnv {
		logger := slog.New(logs.NewCloudLoggingHandler(slog.LevelInfo))
		dbPrefix := ""
		if os.Getenv("STAGE") != "prod" {
			dbPrefix = "test_"
			logger = slog.New(logs.NewCloudLoggingHandler(slog.LevelDebug))
		}
		slog.SetDefault(logger)

		db = repository.NewConnection(context.Background(), conf.FirestoreCredentialFile,
			os.Getenv("OBLIVIATE_PROJECT_ID"), dbPrefix, conf.ProdEnv)
		algorithm = rsa.NewAlgorithm()
	} else {
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})))
		db = mock.StorageMock()
		algorithm = rsa.NewMockAlgorithm()
		slog.Info("Mock DB and encryption started")
	}

	keys, err := crypt.NewKeys(db, &conf, algorithm, true)
	if err != nil {
		slog.Error("error getting keys", logs.Error, err)
	}

	app := app.NewApp(db, &conf, keys)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /variables", handler.ProcessTemplate(&conf, keys.PublicKeyEncoded))
	mux.HandleFunc("POST /save", handler.Save(app))
	mux.HandleFunc("POST /read", handler.Read(app))
	mux.HandleFunc("DELETE /expired", handler.Expired(app))
	mux.HandleFunc("DELETE /delete", handler.Delete(app))
	mux.Handle("GET /", handler.StaticFiles(&conf, true))

	var finalHandler http.Handler = mux
	finalHandler = logs.WithCloudTraceContext(finalHandler)
	finalHandler = gziphandler.GzipHandler(finalHandler)
	if !conf.ProdEnv {
		finalHandler = corsMiddleware(finalHandler)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	slog.Info("Service ready")
	err = http.ListenAndServe(fmt.Sprintf(":%s", port), finalHandler)
	if err != nil {
		slog.Error("Error ListenAndServe", logs.Error, err)
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin == "http://localhost:5173" || origin == "https://localhost:5173" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

