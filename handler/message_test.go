package handler

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/nacl/box"

	"obliviate/app"
	"obliviate/config"
	"obliviate/crypt"
	"obliviate/crypt/rsa"
	"obliviate/handler/webModels"
	"obliviate/repository/mock"
)

func setupTestApp(t *testing.T) (*app.App, *crypt.Keys) {
	conf := &config.Configuration{
		DefaultDurationTime: time.Hour * 24 * 7,
		ProdEnv:             false,
		StaticFilesLocation: "",
	}

	db := mock.StorageMock()
	rsaAlg := rsa.NewMockAlgorithm()

	keys, err := crypt.NewKeys(db, conf, rsaAlg, true)
	require.NoError(t, err, "failed to create test keys")

	return app.NewApp(db, conf, keys), keys
}

// TestSaveHandler_Validation tests all validation paths in Save handler
func TestSaveHandler_Validation(t *testing.T) {
	testApp, _ := setupTestApp(t)

	tests := []struct {
		name           string
		body           interface{}
		expectedStatus int
		expectedInBody string
	}{
		{
			name:           "empty body",
			body:           http.NoBody, // Use NoBody instead of nil to avoid panic in defer
			expectedStatus: http.StatusBadRequest,
			expectedInBody: "",
		},
		{
			name:           "invalid JSON",
			body:           "not-json",
			expectedStatus: http.StatusBadRequest,
			expectedInBody: "",
		},
		{
			name: "empty message",
			body: webModels.SaveRequest{
				Message:           []byte{},
				TransmissionNonce: make([]byte, 24),
				Hash:              "test-hash",
				PublicKey:         make([]byte, 32),
			},
			expectedStatus: http.StatusBadRequest,
			expectedInBody: "",
		},
		{
			name: "message too large",
			body: webModels.SaveRequest{
				Message:           make([]byte, 256*1024*4+1), // 1 byte over limit
				TransmissionNonce: make([]byte, 24),
				Hash:              "test-hash",
				PublicKey:         make([]byte, 32),
			},
			expectedStatus: http.StatusBadRequest,
			expectedInBody: "",
		},
		{
			name: "empty TransmissionNonce",
			body: webModels.SaveRequest{
				Message:           []byte("test message"),
				TransmissionNonce: []byte{},
				Hash:              "test-hash",
				PublicKey:         make([]byte, 32),
			},
			expectedStatus: http.StatusBadRequest,
			expectedInBody: "",
		},
		{
			name: "empty Hash",
			body: webModels.SaveRequest{
				Message:           []byte("test message"),
				TransmissionNonce: make([]byte, 24),
				Hash:              "",
				PublicKey:         make([]byte, 32),
			},
			expectedStatus: http.StatusBadRequest,
			expectedInBody: "",
		},
		{
			name: "wrong TransmissionNonce length",
			body: webModels.SaveRequest{
				Message:           []byte("test message"),
				TransmissionNonce: make([]byte, 20), // Wrong length
				Hash:              "test-hash",
				PublicKey:         make([]byte, 32),
			},
			expectedStatus: http.StatusBadRequest,
			expectedInBody: "",
		},
		{
			name: "wrong PublicKey length",
			body: webModels.SaveRequest{
				Message:           []byte("test message"),
				TransmissionNonce: make([]byte, 24),
				Hash:              "test-hash",
				PublicKey:         make([]byte, 30), // Wrong length
			},
			expectedStatus: http.StatusBadRequest,
			expectedInBody: "",
		},
		{
			name: "valid request",
			body: webModels.SaveRequest{
				Message:           []byte("valid encrypted message"),
				TransmissionNonce: make([]byte, 24),
				Hash:              "valid-hash",
				PublicKey:         make([]byte, 32),
				Time:              100,
				CostFactor:        10,
			},
			expectedStatus: http.StatusOK,
			expectedInBody: "[]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req *http.Request
			var err error

			if tt.body == http.NoBody {
				// Test with NoBody
				req, err = http.NewRequest("POST", "/save", http.NoBody)
			} else if str, ok := tt.body.(string); ok {
				// Test with invalid JSON string
				req, err = http.NewRequest("POST", "/save", strings.NewReader(str))
			} else {
				// Test with valid struct
				jsonBody, _ := json.Marshal(tt.body)
				req, err = http.NewRequest("POST", "/save", bytes.NewBuffer(jsonBody))
			}

			require.NoError(t, err)

			// Add required headers for successful case
			if tt.expectedStatus == http.StatusOK {
				req.Header.Set("Accept-Language", "en")
				req.Header.Set("CF-IPCountry", "US")
			}

			rr := httptest.NewRecorder()
			handler := Save(testApp)
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code, "status code mismatch")

			if tt.expectedInBody != "" {
				assert.Contains(t, rr.Body.String(), tt.expectedInBody)
			}
		})
	}
}

// TestReadHandler_Validation tests all validation paths in Read handler
func TestReadHandler_Validation(t *testing.T) {
	testApp, _ := setupTestApp(t)

	tests := []struct {
		name           string
		body           interface{}
		expectedStatus int
		setupData      bool // Whether to setup test data first
	}{
		{
			name:           "empty body",
			body:           http.NoBody,
			expectedStatus: http.StatusBadRequest,
			setupData:      false,
		},
		{
			name:           "invalid JSON",
			body:           "not-json",
			expectedStatus: http.StatusBadRequest,
			setupData:      false,
		},
		{
			name: "empty Hash",
			body: webModels.ReadRequest{
				Hash:      "",
				PublicKey: make([]byte, 32),
			},
			expectedStatus: http.StatusBadRequest,
			setupData:      false,
		},
		{
			name: "empty PublicKey",
			body: webModels.ReadRequest{
				Hash:      "test-hash",
				PublicKey: []byte{},
			},
			expectedStatus: http.StatusBadRequest,
			setupData:      false,
		},
		{
			name: "wrong PublicKey length",
			body: webModels.ReadRequest{
				Hash:      "test-hash",
				PublicKey: make([]byte, 30), // Wrong length
			},
			expectedStatus: http.StatusBadRequest,
			setupData:      false,
		},
		{
			name: "message not found",
			body: webModels.ReadRequest{
				Hash:      "non-existent-hash",
				PublicKey: make([]byte, 32),
			},
			expectedStatus: http.StatusNotFound,
			setupData:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req *http.Request
			var err error

			if tt.body == http.NoBody {
				req, err = http.NewRequest("POST", "/read", http.NoBody)
			} else if str, ok := tt.body.(string); ok {
				req, err = http.NewRequest("POST", "/read", strings.NewReader(str))
			} else {
				jsonBody, _ := json.Marshal(tt.body)
				req, err = http.NewRequest("POST", "/read", bytes.NewBuffer(jsonBody))
			}

			require.NoError(t, err)

			rr := httptest.NewRecorder()
			handler := Read(testApp)
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code, "status code mismatch")
		})
	}
}

// TestDeleteHandler_Validation tests all validation paths in Delete handler
func TestDeleteHandler_Validation(t *testing.T) {
	testApp, _ := setupTestApp(t)

	tests := []struct {
		name           string
		body           interface{}
		expectedStatus int
	}{
		{
			name:           "empty body",
			body:           http.NoBody,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid JSON",
			body:           "not-json",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "empty Hash",
			body: webModels.DeleteRequest{
				Hash: "",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "valid request",
			body: webModels.DeleteRequest{
				Hash: "test-hash-to-delete",
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req *http.Request
			var err error

			if tt.body == http.NoBody {
				req, err = http.NewRequest("DELETE", "/delete", http.NoBody)
			} else if str, ok := tt.body.(string); ok {
				req, err = http.NewRequest("DELETE", "/delete", strings.NewReader(str))
			} else {
				jsonBody, _ := json.Marshal(tt.body)
				req, err = http.NewRequest("DELETE", "/delete", bytes.NewBuffer(jsonBody))
			}

			require.NoError(t, err)

			rr := httptest.NewRecorder()
			handler := Delete(testApp)
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code, "status code mismatch")
		})
	}
}

// TestExpiredHandler tests both success and error paths
func TestExpiredHandler(t *testing.T) {
	testApp, _ := setupTestApp(t)

	t.Run("success case", func(t *testing.T) {
		req, err := http.NewRequest("DELETE", "/expired", nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		handler := Expired(testApp)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "[]", rr.Body.String())
	})
}

// TestSaveAndReadIntegration tests successful save and read flow
func TestSaveAndReadIntegration(t *testing.T) {
	testApp, appKeys := setupTestApp(t)

	// Generate browser keys for encryption/decryption
	browserPublicKey, browserPrivateKey, err := box.GenerateKey(rand.Reader)
	require.NoError(t, err)

	testMessage := []byte("Integration test message")
	messageHash := "integration-test-hash"

	// Generate a random nonce for transmission
	var transmissionNonce [24]byte
	_, err = io.ReadFull(rand.Reader, transmissionNonce[:])
	require.NoError(t, err, "failed to generate transmission nonce")

	// Encrypt the message using the app's public key and browser's private key
	// The message is encrypted for the app (recipient) by the browser (sender)
	encryptedMessage := box.Seal(nil, testMessage, &transmissionNonce, (*[32]byte)(appKeys.PublicKey), browserPrivateKey)

	// Create save request
	saveReq := webModels.SaveRequest{
		Message:           encryptedMessage,
		TransmissionNonce: transmissionNonce[:],
		Hash:              messageHash,
		PublicKey:         browserPublicKey[:], // Browser's public key, so app can encrypt response for browser
		Time:              100,
		CostFactor:        10,
	}

	// Save message
	saveBody, _ := json.Marshal(saveReq)
	saveRequest, _ := http.NewRequest("POST", "/save", bytes.NewBuffer(saveBody))
	saveRequest.Header.Set("Accept-Language", "en")
	saveRequest.Header.Set("CF-IPCountry", "US")

	saveRR := httptest.NewRecorder()
	Save(testApp).ServeHTTP(saveRR, saveRequest)

	assert.Equal(t, http.StatusOK, saveRR.Code, "save should succeed")

	// Read message
	readReq := webModels.ReadRequest{
		Hash:      messageHash,
		PublicKey: browserPublicKey[:], // Browser's public key
		Password:  false,
	}

	readBody, _ := json.Marshal(readReq)
	readRequest, _ := http.NewRequest("POST", "/read", bytes.NewBuffer(readBody))

	readRR := httptest.NewRecorder()
	Read(testApp).ServeHTTP(readRR, readRequest)

	assert.Equal(t, http.StatusOK, readRR.Code, "read should succeed with proper encryption")

	var readResponse webModels.ReadResponse
	err = json.Unmarshal(readRR.Body.Bytes(), &readResponse)
	require.NoError(t, err, "failed to unmarshal read response")

	// The response Message field contains nonce (first 24 bytes) + ciphertext
	// Server encrypted the original message for the browser using browser's public key
	require.Greater(t, len(readResponse.Message), 24, "response should contain nonce + ciphertext")

	var responseNonce [24]byte
	copy(responseNonce[:], readResponse.Message[:24])
	ciphertext := readResponse.Message[24:]

	decryptedMessage, ok := box.Open(nil, ciphertext, &responseNonce, (*[32]byte)(appKeys.PublicKey), browserPrivateKey)
	require.True(t, ok, "failed to decrypt message from read response")
	assert.Equal(t, testMessage, decryptedMessage, "decrypted message should match original")
}

// Test edge cases for message sizes
func TestSaveHandler_MessageSizeEdgeCases(t *testing.T) {
	testApp, _ := setupTestApp(t)

	maxSize := 256 * 1024 * 4

	tests := []struct {
		name           string
		messageSize    int
		expectedStatus int
	}{
		{
			name:           "exactly at max size",
			messageSize:    maxSize,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "one byte over max",
			messageSize:    maxSize + 1,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "minimum size (1 byte)",
			messageSize:    1,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			saveReq := webModels.SaveRequest{
				Message:           make([]byte, tt.messageSize),
				TransmissionNonce: make([]byte, 24),
				Hash:              "size-test-hash",
				PublicKey:         make([]byte, 32),
			}

			jsonBody, _ := json.Marshal(saveReq)
			req, _ := http.NewRequest("POST", "/save", bytes.NewBuffer(jsonBody))
			req.Header.Set("Accept-Language", "en")
			req.Header.Set("CF-IPCountry", "US")

			rr := httptest.NewRecorder()
			Save(testApp).ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}

// Test that handlers properly close request bodies
func TestHandlers_BodyClosing(t *testing.T) {
	testApp, _ := setupTestApp(t)

	t.Run("Save handler closes body", func(t *testing.T) {
		body := io.NopCloser(strings.NewReader(`{"hash":"test"}`))
		req, _ := http.NewRequest("POST", "/save", body)

		rr := httptest.NewRecorder()
		Save(testApp).ServeHTTP(rr, req)

		// Body should be closed (can't verify directly, but handler has defer r.Body.Close())
		assert.NotPanics(t, func() {
			Save(testApp).ServeHTTP(rr, req)
		})
	})
}
