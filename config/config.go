package config

import (
	"obliviate/interfaces/store"
	"time"
)

type Configuration struct {
	Db                      store.Connection
	DefaultDurationTime     time.Duration
	ProdEnv                 bool
	MasterKey               string
	KmsCredentialFile       string
	FirestoreCredentialFile string
}
