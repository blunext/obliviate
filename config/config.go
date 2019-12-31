package config

import (
	"time"
)

type Configuration struct {
	DefaultDurationTime     time.Duration
	ProdEnv                 bool
	MasterKey               string
	KmsCredentialFile       string
	FirestoreCredentialFile string
}
