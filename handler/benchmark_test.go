package handler

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/nacl/box"

	"obliviate/app"
	"obliviate/config"
	"obliviate/crypt"
	"obliviate/crypt/rsa"
	"obliviate/handler/webModels"
	"obliviate/repository/mock"
)

func setupBenchmarkApp(b *testing.B) *app.App {
	conf := &config.Configuration{
		DefaultDurationTime: time.Hour,
		ProdEnv:             false,
	}
	db := mock.StorageMock()
	rsaAlg := rsa.NewMockAlgorithm()
	keys, err := crypt.NewKeys(db, conf, rsaAlg, true)
	require.NoError(b, err)

	return app.NewApp(db, conf, keys)
}

func BenchmarkSaveHandler(b *testing.B) {
	application := setupBenchmarkApp(b)
	handlerFunc := Save(application)

	senderPub, _, _ := box.GenerateKey(rand.Reader)
	message := make([]byte, 1024) // 1KB message
	rand.Read(message)

	reqBody := webModels.SaveRequest{
		Message:           message,
		TransmissionNonce: make([]byte, 24),
		Hash:              "bench-hash",
		PublicKey:         senderPub[:],
		Time:              100,
		CostFactor:        10,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/save", bytes.NewReader(bodyBytes))
		w := httptest.NewRecorder()
		handlerFunc.ServeHTTP(w, req)
	}
}

func BenchmarkReadHandler(b *testing.B) {
	conf := &config.Configuration{DefaultDurationTime: time.Hour}
	db := mock.StorageMock()
	rsaAlg := rsa.NewMockAlgorithm()
	keys, _ := crypt.NewKeys(db, conf, rsaAlg, true)
	appInstance := app.NewApp(db, conf, keys)

	saveHandler := Save(appInstance)
	readHandler := Read(appInstance)

	// Prepare a message
	senderPub, senderPriv, _ := box.GenerateKey(rand.Reader)
	recipientPub, _, _ := box.GenerateKey(rand.Reader)

	// We need to encrypt it properly so Read can decrypt it
	// Read expects: BoxOpen(storedMsg, senderPub, storedNonce) -> decrypted
	// So we store: BoxSeal(nil, decrypted, nonce, serverPub, senderPriv)
	// Wait, we need server keys to encrypt for server.
	// setupBenchmarkApp doesn't return keys.
	// But we can cheat: we can use a mock DB that returns what we want,
	// OR we can just use the public flow if we had keys.

	// Let's just use the public flow. We need keys.
	// I'll modify setupBenchmarkApp to return keys or just create them here.

	message := []byte("benchmark payload")
	var nonce [24]byte
	rand.Read(nonce[:])

	encryptedMsg := box.Seal(nil, message, &nonce, keys.PublicKey, senderPriv)

	saveReq := webModels.SaveRequest{
		Message:           encryptedMsg,
		TransmissionNonce: nonce[:],
		Hash:              "bench-read-hash",
		PublicKey:         senderPub[:],
	}
	saveBody, _ := json.Marshal(saveReq)

	readReq := webModels.ReadRequest{
		Hash:      "bench-read-hash",
		PublicKey: recipientPub[:],
	}
	readBody, _ := json.Marshal(readReq)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		// Save message again because Read deletes it
		reqSave := httptest.NewRequest("POST", "/save", bytes.NewReader(saveBody))
		wSave := httptest.NewRecorder()
		saveHandler.ServeHTTP(wSave, reqSave)
		b.StartTimer()

		req := httptest.NewRequest("POST", "/read", bytes.NewReader(readBody))
		w := httptest.NewRecorder()
		readHandler.ServeHTTP(w, req)
	}
}

func BenchmarkCrypto_BoxSeal(b *testing.B) {
	pub, priv, _ := box.GenerateKey(rand.Reader)
	message := make([]byte, 1024)
	rand.Read(message)
	var nonce [24]byte
	rand.Read(nonce[:])

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		box.Seal(nil, message, &nonce, pub, priv)
	}
}

func BenchmarkCrypto_BoxOpen(b *testing.B) {
	pub, priv, _ := box.GenerateKey(rand.Reader)
	message := make([]byte, 1024)
	rand.Read(message)
	var nonce [24]byte
	rand.Read(nonce[:])

	encrypted := box.Seal(nil, message, &nonce, pub, priv)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = box.Open(nil, encrypted, &nonce, pub, priv)
	}
}
