package main

import (
	"context"
	"embed"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	_ "golang.org/x/crypto/x509roots/fallback"

	"obliviate/logs"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

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
//go:embed web/build/*
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

	r := chi.NewRouter()

	if !conf.ProdEnv {
		r.Use(cors.Handler(cors.Options{
			AllowedOrigins: []string{"https://localhost:5173", "http://localhost:5173"},
			AllowedMethods: []string{"GET", "POST", "DELETE", "OPTIONS"},
			AllowedHeaders: []string{"Content-Type"},
		}))
	}

	compressor := middleware.NewCompressor(5, "text/html", "text/javascript", "application/javascript", "text/css", "image/x-icon", "text/plain", "application/json")
	r.Use(compressor.Handler)
	r.Use(logs.WithCloudTraceContext)

	r.Get("/*", handler.StaticFiles(&conf, true))
	r.Get("/variables", handler.ProcessTemplate(&conf, keys.PublicKeyEncoded))
	r.Post("/save", handler.Save(app))
	r.Post("/read", handler.Read(app))
	r.Delete("/expired", handler.Expired(app))
	r.Delete("/delete", handler.Delete(app))

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	slog.Info("Service ready")
	err = http.ListenAndServe(fmt.Sprintf(":%s", port), r)
	if err != nil {
		slog.Error("Error ListenAndServe", logs.Error, err)
	}
}
