package crypt

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"obliviate/config"
	"obliviate/crypt/rsa"
	"obliviate/repository"
	"obliviate/repository/mock"
)

var conf *config.Configuration
var db repository.DataBase

func init() {
	conf = &config.Configuration{
		DefaultDurationTime:     time.Hour * 24 * 7,
		ProdEnv:                 os.Getenv("ENV") == "PROD",
		MasterKey:               os.Getenv("HSM_MASTER_KEY"),
		KmsCredentialFile:       os.Getenv("KMS_CREDENTIAL_FILE"),
		FirestoreCredentialFile: os.Getenv("FIRESTORE_CREDENTIAL_FILE"),
	}
	// conf.Db = repository.Connect(context.Background(), "test")
	db = mock.StorageMock()
}

func TestKeysGenerationAndStorage(t *testing.T) {

	rsa := rsa.NewMockAlgorithm()
	// rsa := rsa.NewAlgorithm()

	keys, err := NewKeys(db, conf, rsa, true)
	assert.NoError(t, err, "should not be error")

	pubKey := keys.PublicKeyEncoded

	var priv [32]byte
	//nolint:gosimple
	var pub [32]byte
	pub = *keys.PublicKey
	priv = *keys.PrivateKey

	keys, err = NewKeys(db, conf, rsa, true)
	assert.NoError(t, err, "should not be error")

	assert.Equal(t, pubKey, keys.PublicKeyEncoded, "private keys should be the same")
	assert.Equal(t, priv, *keys.PrivateKey, "private keys should be the same")
	assert.Equal(t, pub, *keys.PublicKey, "public keys should be the same")

}
