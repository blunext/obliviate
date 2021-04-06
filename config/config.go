package config

import (
	"io/fs"
	"time"
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
