package config

import (
	"io/fs"
	"time"
)

type contextKey string

var (
	CountryCode    = contextKey("country-code")
	AcceptLanguage = contextKey("accept-language")
)

type Configuration struct {
	DefaultDurationTime     time.Duration
	ProdEnv                 bool
	MasterKey               string
	KmsCredentialFile       string
	FirestoreCredentialFile string
	StaticFilesLocation     string
	EmbededStaticFiles      fs.FS
}
