package main

import (
	"context"
	"embed"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/sirupsen/logrus"

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
//go:embed web/build
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
		initLogrus(logrus.InfoLevel)
		db = repository.NewConnection(context.Background(), conf.FirestoreCredentialFile,
			os.Getenv("OBLIVIATE_PROJECT_ID"), conf.ProdEnv)
		algorithm = rsa.NewAlgorithm()
	} else {
		initLogrus(logrus.TraceLevel)
		db = mock.StorageMock()
		algorithm = rsa.NewMockAlgorithm()
		logrus.Info("Mock DB and encryption started")
	}

	keys, err := crypt.NewKeys(db, &conf, algorithm, true)
	if err != nil {
		logrus.Panicf("error getting keys, err: %v", err)
	}

	app := app.NewApp(db, &conf, keys)

	r := chi.NewRouter()
	compressor := middleware.NewCompressor(5, "text/html", "text/javascript", "application/javascript", "text/css", "image/x-icon", "text/plain", "application/json")
	r.Use(compressor.Handler)

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

	logrus.Info("Service ready")
	err = http.ListenAndServe(fmt.Sprintf(":%s", port), r)
	if err != nil {
		logrus.Errorf("Error ListenAndServe: %v", err)
	}
}

func initLogrus(level logrus.Level) {

	if os.Getenv("ENV") == "PROD" {
		logrus.SetFormatter(&logrus.JSONFormatter{
			FieldMap: logrus.FieldMap{
				logrus.FieldKeyTime:  "time",
				logrus.FieldKeyLevel: "severity",
				logrus.FieldKeyMsg:   "message",
			},
		})
	} else {
		logrus.SetFormatter(&logrus.TextFormatter{
			TimestampFormat: "02-01-2006 15:04:05",
			FullTimestamp:   true,
		})
	}
	logrus.SetLevel(level)
}
