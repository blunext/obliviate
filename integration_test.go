package main

import (
	"bytes"
	"crypto/rand"
	"embed"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/nacl/box"

	"obliviate/app"
	"obliviate/config"
	"obliviate/crypt"
	"obliviate/crypt/rsa"
	"obliviate/handler"
	"obliviate/handler/webModels"
	"obliviate/repository/mock"
)

//go:embed handler/testdata/variables.json
var testStaticFiles embed.FS

func setupIntegrationApp(t *testing.T) (*app.App, *chi.Mux, *crypt.Keys) {
	conf := &config.Configuration{
		DefaultDurationTime: time.Hour * 24 * 7,
		ProdEnv:             false,
		EmbededStaticFiles:  testStaticFiles,
	}

	db := mock.StorageMock()
	rsaAlg := rsa.NewMockAlgorithm()

	keys, err := crypt.NewKeys(db, conf, rsaAlg, true)
	require.NoError(t, err)

	application := app.NewApp(db, conf, keys)

	r := chi.NewRouter()
	r.Post("/save", handler.Save(application))
	r.Post("/read", handler.Read(application))
	r.Delete("/delete", handler.Delete(application))
	r.Delete("/expired", handler.Expired(application))

	return application, r, keys
}

func TestEndToEnd_MessageLifecycle(t *testing.T) {
	_, router, serverKeys := setupIntegrationApp(t)

	// 1. Setup Client Keys
	senderPub, senderPriv, err := box.GenerateKey(rand.Reader)
	require.NoError(t, err)

	// 2. Prepare Message
	originalMessage := []byte("This is a secret message for integration test")

	// Encrypt message for Server
	var nonce [24]byte
	_, err = rand.Read(nonce[:])
	require.NoError(t, err)

	// box.Seal appends the encrypted message to the first argument.
	// We want just the ciphertext for the Message field, as Nonce is sent separately.
	encryptedMessage := box.Seal(nil, originalMessage, &nonce, serverKeys.PublicKey, senderPriv)

	hash := "integration-lifecycle-hash"

	saveReq := webModels.SaveRequest{
		Message:           encryptedMessage,
		TransmissionNonce: nonce[:], // Use the REAL nonce
		Hash:              hash,
		PublicKey:         senderPub[:],
		Time:              100,
		CostFactor:        10,
	}

	// 3. SAVE
	saveBody, _ := json.Marshal(saveReq)
	reqSave := httptest.NewRequest("POST", "/save", bytes.NewBuffer(saveBody))
	wSave := httptest.NewRecorder()
	router.ServeHTTP(wSave, reqSave)

	require.Equal(t, http.StatusOK, wSave.Code, "Save should succeed")

	// 4. READ
	// To read, we need to provide a PublicKey. The server will encrypt the message for this key.
	recipientPub, recipientPriv, err := box.GenerateKey(rand.Reader)
	require.NoError(t, err)

	readReq := webModels.ReadRequest{
		Hash:      hash,
		PublicKey: recipientPub[:],
	}

	readBody, _ := json.Marshal(readReq)
	reqRead := httptest.NewRequest("POST", "/read", bytes.NewBuffer(readBody))
	wRead := httptest.NewRecorder()
	router.ServeHTTP(wRead, reqRead)

	require.Equal(t, http.StatusOK, wRead.Code, "Read should succeed")

	var readResp webModels.ReadResponse
	err = json.NewDecoder(wRead.Body).Decode(&readResp)
	require.NoError(t, err)

	// 5. Decrypt response
	// The server encrypted the DECRYPTED message (originalMessage) using recipientPub and serverPriv.
	// We decrypt using recipientPriv and serverPub.

	// Extract nonce from response (first 24 bytes)
	require.True(t, len(readResp.Message) > 24)
	var respNonce [24]byte
	copy(respNonce[:], readResp.Message[:24])
	cipherText := readResp.Message[24:]

	decryptedMessage, ok := box.Open(nil, cipherText, &respNonce, serverKeys.PublicKey, recipientPriv)
	require.True(t, ok, "Failed to decrypt response from server")
	assert.Equal(t, originalMessage, decryptedMessage, "Decrypted message should match original")

	// 6. DELETE
	delReq := webModels.DeleteRequest{Hash: hash}
	delBody, _ := json.Marshal(delReq)
	reqDel := httptest.NewRequest("DELETE", "/delete", bytes.NewBuffer(delBody))
	wDel := httptest.NewRecorder()
	router.ServeHTTP(wDel, reqDel)

	require.Equal(t, http.StatusOK, wDel.Code, "Delete should succeed")

	// 7. READ AGAIN (Should fail)
	reqRead2 := httptest.NewRequest("POST", "/read", bytes.NewBuffer(readBody))
	wRead2 := httptest.NewRecorder()
	router.ServeHTTP(wRead2, reqRead2)

	assert.Equal(t, http.StatusNotFound, wRead2.Code, "Read after delete should return 404")
}

func TestEndToEnd_Expiration(t *testing.T) {
	appInstance, router, _ := setupIntegrationApp(t)

	hash := "integration-expiration-hash"
	senderPub, _, _ := box.GenerateKey(rand.Reader)

	// 1. Save with short TTL?
	// The `SaveRequest` doesn't have TTL. It has `Time` and `CostFactor` for PoW?
	// Or `ValidTo` is calculated?
	// `ProcessSave` calls `app.repo.SaveMessage`.
	// `MessageModel` has `ValidTo`.
	// `ProcessSave` sets `ValidTo = time.Now().Add(app.Config.DefaultDurationTime)`.
	// It seems we cannot control TTL from API request in `ProcessSave` implementation?
	// Let's check `app/service.go`.
	// `ProcessSave`:
	// m := model.MessageModel{ ... ValidTo: time.Now().Add(app.Config.DefaultDurationTime) ... }

	// So we can't easily test expiration unless we mock time or change DefaultDurationTime.
	// `setupIntegrationApp` sets `DefaultDurationTime`.
	// We should allow customizing it or modify the app config after setup.

	// Let's modify setup to allow config tweaks or just modify it here since we have access to `app`.

	// But `setupIntegrationApp` returns `*app.App`.
	// We can modify `app.Config.DefaultDurationTime`.

	// appInstance, router := setupIntegrationApp(t)
	appInstance.Config.DefaultDurationTime = 100 * time.Millisecond // Short TTL

	saveReq := webModels.SaveRequest{
		Message:           []byte("msg"),
		TransmissionNonce: make([]byte, 24),
		Hash:              hash,
		PublicKey:         senderPub[:],
	}

	saveBody, _ := json.Marshal(saveReq)
	reqSave := httptest.NewRequest("POST", "/save", bytes.NewBuffer(saveBody))
	wSave := httptest.NewRecorder()
	router.ServeHTTP(wSave, reqSave)
	require.Equal(t, http.StatusOK, wSave.Code)

	// 2. Wait for expiration
	time.Sleep(200 * time.Millisecond)

	// 3. Trigger Expired cleanup
	reqExpired := httptest.NewRequest("DELETE", "/expired", nil)
	wExpired := httptest.NewRecorder()
	router.ServeHTTP(wExpired, reqExpired)
	require.Equal(t, http.StatusOK, wExpired.Code)

	// 4. Read (Should be gone)
	readReq := webModels.ReadRequest{
		Hash:      hash,
		PublicKey: senderPub[:],
	}
	readBody, _ := json.Marshal(readReq)
	reqRead := httptest.NewRequest("POST", "/read", bytes.NewBuffer(readBody))
	wRead := httptest.NewRecorder()
	router.ServeHTTP(wRead, reqRead)

	assert.Equal(t, http.StatusNotFound, wRead.Code, "Expired message should be deleted")
}
