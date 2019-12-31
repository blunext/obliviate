package crypt

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/nacl/box"
	"io"
	"obliviate/config"
	"obliviate/crypt/rsa"
	"obliviate/interfaces/store"
)

type Keys struct {
	PublicKey        *[32]byte
	PrivateKey       *[32]byte
	PublicKeyEncoded string
}

func NewKeys(db store.Connection, conf *config.Configuration, algorithm rsa.RSA) (*Keys, error) {

	k := Keys{}

	encrypted, err := db.GetEncryptedKeys(context.Background())
	if err != nil {
		return nil, fmt.Errorf("error retreaving keys from DB: %v", err)
	}

	if encrypted != nil {
		// decrypt Keys
		decrypted, err := algorithm.DecryptRSA(conf, encrypted)
		if err != nil {
			return nil, fmt.Errorf("error decryting keys: %v", err)
		}
		k.PublicKey = new([32]byte)
		k.PrivateKey = new([32]byte)

		copy(k.PublicKey[:], decrypted[:32])
		copy(k.PrivateKey[:], decrypted[32:])

		logrus.Trace("encryption keys fetched and decrypted by master key")

	} else {
		// generate Keys
		k.PublicKey, k.PrivateKey, err = box.GenerateKey(rand.Reader)
		if err != nil {
			return nil, fmt.Errorf("error generating keys: %v", err)
		}
		both := append(k.PublicKey[:], k.PrivateKey[:]...)

		// encrypt Keys
		encrypted, err = algorithm.EncryptRSA(conf, both)
		if err != nil {
			return nil, fmt.Errorf("error encrypting keys: %v", err)
		}

		// store crypted Keys
		err = db.SaveEncryptedKeys(context.Background(), encrypted)
		if err != nil {
			return nil, fmt.Errorf("error storing keys into DB: %v", err)
		}

		logrus.Trace("encryption keys generate, encrypted by master key, stored in DB")
	}

	k.PublicKeyEncoded = base64.StdEncoding.EncodeToString(k.PublicKey[:])
	logrus.Info("encryption keys are ready")

	return &k, nil
}

func (keys *Keys) BoxOpen(encrypted []byte, senderPublicKey *[32]byte, decryptNonce *[24]byte) ([]byte, error) {

	decrypted, ok := box.Open(nil, encrypted, decryptNonce, senderPublicKey, keys.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("cannot make box open")
	}
	return decrypted, nil
}

func (keys *Keys) BoxSeal(msg []byte, recipientPublicKey *[32]byte) ([]byte, error) {

	var nonce [24]byte
	var err error
	if nonce, err = keys.GenerateNonce(); err != nil {
		return nil, err
	}

	encrypted := box.Seal(nonce[:], msg, &nonce, recipientPublicKey, keys.PrivateKey)
	// nonce is already included in first 24 bytes of encrypted message
	return encrypted, nil
}

func (keys *Keys) GenerateNonce() ([24]byte, error) {
	var nonce [24]byte
	if _, err := io.ReadFull(rand.Reader, nonce[:]); err != nil {
		return nonce, fmt.Errorf("cannot generate nounce, err: %w", err)
	}
	return nonce, nil
}
