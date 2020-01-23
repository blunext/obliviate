package main

import (
	"context"
	"fmt"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/sirupsen/logrus"
	"net/http"
	"obliviate/app"
	"obliviate/config"
	"obliviate/crypt"
	"obliviate/crypt/rsa"
	"obliviate/interfaces/handler"
	"obliviate/interfaces/store"
	"obliviate/interfaces/store/mock"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	messageCollectionName = "messages"
	messageDurationTime   = time.Hour * 24 * 7 * 4
)

func main() {

	conf := config.Configuration{
		DefaultDurationTime:     messageDurationTime,
		ProdEnv:                 os.Getenv("ENV") == "PROD",
		MasterKey:               os.Getenv("HSM_MASTER_KEY"),
		KmsCredentialFile:       os.Getenv("KMS_CREDENTIAL_FILE"),
		FirestoreCredentialFile: os.Getenv("FIRESTORE_CREDENTIAL_FILE"),
	}

	var algorithm rsa.RSA
	var db store.Connection

	if conf.ProdEnv {
		initLogrus(logrus.DebugLevel)
		db = store.NewConnection(context.Background(), messageCollectionName, conf.FirestoreCredentialFile,
			os.Getenv("OBLIVIATE_PROJECT_ID"), conf.ProdEnv)
		algorithm = rsa.NewAlgorithm()
	} else {
		initLogrus(logrus.TraceLevel)
		db = mock.StorageMock()
		algorithm = rsa.NewMockAlgorithm()
	}

	keys, err := crypt.NewKeys(db, &conf, algorithm)
	if err != nil {
		logrus.Panicf("error getting keys, err: %v", err)
	}

	app := app.NewApp(db, &conf, keys)

	r := chi.NewRouter()
	r.Use(middleware.DefaultCompress)

	r.Get("/", handler.ProcessTemplate(&conf, keys.PublicKeyEncoded))
	r.Post("/save", handler.Save(app))
	r.Post("/read", handler.Read(app))
	r.Delete("/expired", handler.Expired(app))

	workDir, _ := os.Getwd()
	filesDir := filepath.Join(workDir, "static")
	FileServer(r, "/static", http.Dir(filesDir))

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	logrus.Info("Server starts")
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

func FileServer(r chi.Router, path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit URL parameters.")
	}

	fs := http.StripPrefix(path, http.FileServer(root))

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", 301).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, func(w http.ResponseWriter, r *http.Request) {
		fs.ServeHTTP(w, r)
	})
}
