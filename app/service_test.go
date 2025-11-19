package app

import (
	"context"
	"crypto/rand"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/nacl/box"

	"obliviate/config"
	"obliviate/crypt"
	"obliviate/crypt/rsa"
	"obliviate/handler/webModels"
	"obliviate/repository"
	"obliviate/repository/mock"
)

var (
	testConf *config.Configuration
	testDB   repository.DataBase
	testKeys *crypt.Keys
	testApp  *App
)

func setupTest(t *testing.T) {
	testConf = &config.Configuration{
		DefaultDurationTime:     time.Hour * 24 * 7,
		ProdEnv:                 false,
		MasterKey:               os.Getenv("HSM_MASTER_KEY"),
		KmsCredentialFile:       os.Getenv("KMS_CREDENTIAL_FILE"),
		FirestoreCredentialFile: os.Getenv("FIRESTORE_CREDENTIAL_FILE"),
	}

	testDB = mock.StorageMock()
	rsaAlg := rsa.NewMockAlgorithm()

	var err error
	testKeys, err = crypt.NewKeys(testDB, testConf, rsaAlg, true)
	require.NoError(t, err, "failed to create test keys")

	testApp = NewApp(testDB, testConf, testKeys)
}

func TestNewApp(t *testing.T) {
	setupTest(t)

	assert.NotNil(t, testApp)
	assert.Equal(t, testConf, testApp.Config)
	assert.NotNil(t, testApp.keys)
	assert.NotNil(t, testApp.db)
}

func TestProcessSave(t *testing.T) {
	setupTest(t)

	tests := []struct {
		name        string
		setupCtx    func() context.Context
		request     webModels.SaveRequest
		wantErr     bool
		errContains string
	}{
		{
			name: "valid message save",
			setupCtx: func() context.Context {
				ctx := context.Background()
				ctx = context.WithValue(ctx, config.CountryCode, "PL")
				return ctx
			},
			request: webModels.SaveRequest{
				Message:           []byte("encrypted message content"),
				TransmissionNonce: make([]byte, 24),
				Hash:              "test-hash-123",
				PublicKey:         make([]byte, 32),
				Time:              100,
				CostFactor:        10,
			},
			wantErr: false,
		},
		{
			name: "save with special characters in hash",
			setupCtx: func() context.Context {
				ctx := context.Background()
				ctx = context.WithValue(ctx, config.CountryCode, "US")
				return ctx
			},
			request: webModels.SaveRequest{
				Message:           []byte("test message"),
				TransmissionNonce: make([]byte, 24),
				Hash:              "hash-with-special/chars?test=123",
				PublicKey:         make([]byte, 32),
				Time:              50,
				CostFactor:        5,
			},
			wantErr: false,
		},
		{
			name: "save with empty country code",
			setupCtx: func() context.Context {
				ctx := context.Background()
				ctx = context.WithValue(ctx, config.CountryCode, "")
				return ctx
			},
			request: webModels.SaveRequest{
				Message:           []byte("message without country"),
				TransmissionNonce: make([]byte, 24),
				Hash:              "hash-no-country",
				PublicKey:         make([]byte, 32),
				Time:              75,
				CostFactor:        8,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setupCtx()
			err := testApp.ProcessSave(ctx, tt.request)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)

				// Verify message was saved to database
				// Give goroutine time to increment counter
				time.Sleep(10 * time.Millisecond)
			}
		})
	}
}

func TestProcessRead(t *testing.T) {
	setupTest(t)

	// First, create and save a message
	browserPublicKey, browserPrivateKey, err := box.GenerateKey(rand.Reader)
	require.NoError(t, err)

	testMessage := []byte("Secret test message content")

	// Encrypt the message using server's public key
	nonce, err := testKeys.GenerateNonce()
	require.NoError(t, err)

	encrypted := box.Seal(nonce[:], testMessage, &nonce, testKeys.PublicKey, browserPrivateKey)

	testHash := "test-read-hash"
	ctx := context.WithValue(context.Background(), config.CountryCode, "PL")

	saveReq := webModels.SaveRequest{
		Message:           encrypted[24:], // without nonce
		TransmissionNonce: nonce[:],
		Hash:              testHash,
		PublicKey:         browserPublicKey[:],
		Time:              100,
		CostFactor:        10,
	}

	err = testApp.ProcessSave(ctx, saveReq)
	require.NoError(t, err)

	tests := []struct {
		name           string
		request        webModels.ReadRequest
		wantErr        bool
		errContains    string
		wantNilMessage bool
		wantCostFactor int
	}{
		{
			name: "read existing message without password",
			request: webModels.ReadRequest{
				Hash:      testHash,
				PublicKey: make([]byte, 32),
				Password:  false,
			},
			wantErr:        false,
			wantNilMessage: false,
			wantCostFactor: 10,
		},
		{
			name: "read non-existent message",
			request: webModels.ReadRequest{
				Hash:      "non-existent-hash",
				PublicKey: make([]byte, 32),
				Password:  false,
			},
			wantErr:        false,
			wantNilMessage: true,
			wantCostFactor: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Generate new recipient keys for each test
			recipientPubKey, _, err := box.GenerateKey(rand.Reader)
			require.NoError(t, err)
			tt.request.PublicKey = recipientPubKey[:]

			encrypted, costFactor, err := testApp.ProcessRead(context.Background(), tt.request)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantCostFactor, costFactor)

				if tt.wantNilMessage {
					assert.Nil(t, encrypted)
				} else {
					assert.NotNil(t, encrypted)
					assert.Greater(t, len(encrypted), 24, "encrypted message should include nonce")
				}
			}
		})
	}
}

func TestProcessRead_WithPassword(t *testing.T) {
	setupTest(t)

	// Create a password-protected message
	browserPublicKey, browserPrivateKey, err := box.GenerateKey(rand.Reader)
	require.NoError(t, err)

	testMessage := []byte("Password protected message")
	nonce, err := testKeys.GenerateNonce()
	require.NoError(t, err)

	encrypted := box.Seal(nonce[:], testMessage, &nonce, testKeys.PublicKey, browserPrivateKey)

	testHash := "password-hash"
	ctx := context.WithValue(context.Background(), config.CountryCode, "PL")

	saveReq := webModels.SaveRequest{
		Message:           encrypted[24:],
		TransmissionNonce: nonce[:],
		Hash:              testHash,
		PublicKey:         browserPublicKey[:],
		Time:              100,
		CostFactor:        15,
	}

	err = testApp.ProcessSave(ctx, saveReq)
	require.NoError(t, err)

	// Read with password=true (message should NOT be deleted)
	recipientPubKey, _, err := box.GenerateKey(rand.Reader)
	require.NoError(t, err)

	readReq := webModels.ReadRequest{
		Hash:      testHash,
		PublicKey: recipientPubKey[:],
		Password:  true,
	}

	_, _, err = testApp.ProcessRead(context.Background(), readReq)
	assert.NoError(t, err)

	// Wait for potential deletion goroutine
	time.Sleep(10 * time.Millisecond)

	// Try reading again - should still exist
	recipientPubKey2, _, err := box.GenerateKey(rand.Reader)
	require.NoError(t, err)

	readReq2 := webModels.ReadRequest{
		Hash:      testHash,
		PublicKey: recipientPubKey2[:],
		Password:  true,
	}

	encrypted2, costFactor, err := testApp.ProcessRead(context.Background(), readReq2)
	assert.NoError(t, err)
	assert.NotNil(t, encrypted2, "message should still exist after password-protected read")
	assert.Equal(t, 15, costFactor)
}

func TestProcessRead_AutoDelete(t *testing.T) {
	setupTest(t)

	// Create a regular message (no password)
	browserPublicKey, browserPrivateKey, err := box.GenerateKey(rand.Reader)
	require.NoError(t, err)

	testMessage := []byte("Auto-delete message")
	nonce, err := testKeys.GenerateNonce()
	require.NoError(t, err)

	encrypted := box.Seal(nonce[:], testMessage, &nonce, testKeys.PublicKey, browserPrivateKey)

	testHash := "auto-delete-hash"
	ctx := context.WithValue(context.Background(), config.CountryCode, "US")

	saveReq := webModels.SaveRequest{
		Message:           encrypted[24:],
		TransmissionNonce: nonce[:],
		Hash:              testHash,
		PublicKey:         browserPublicKey[:],
		Time:              100,
		CostFactor:        10,
	}

	err = testApp.ProcessSave(ctx, saveReq)
	require.NoError(t, err)

	// Read with password=false (message SHOULD be deleted)
	recipientPubKey, _, err := box.GenerateKey(rand.Reader)
	require.NoError(t, err)

	readReq := webModels.ReadRequest{
		Hash:      testHash,
		PublicKey: recipientPubKey[:],
		Password:  false,
	}

	_, _, err = testApp.ProcessRead(context.Background(), readReq)
	assert.NoError(t, err)

	// Wait for deletion goroutine to complete
	time.Sleep(50 * time.Millisecond)

	// Try reading again - should NOT exist
	recipientPubKey2, _, err := box.GenerateKey(rand.Reader)
	require.NoError(t, err)

	readReq2 := webModels.ReadRequest{
		Hash:      testHash,
		PublicKey: recipientPubKey2[:],
		Password:  false,
	}

	encrypted2, _, err := testApp.ProcessRead(context.Background(), readReq2)
	assert.NoError(t, err)
	assert.Nil(t, encrypted2, "message should be deleted after non-password read")
}

func TestProcessDelete(t *testing.T) {
	setupTest(t)

	// Create a message first
	browserPublicKey, browserPrivateKey, err := box.GenerateKey(rand.Reader)
	require.NoError(t, err)

	testMessage := []byte("Message to be deleted")
	nonce, err := testKeys.GenerateNonce()
	require.NoError(t, err)

	encrypted := box.Seal(nonce[:], testMessage, &nonce, testKeys.PublicKey, browserPrivateKey)

	testHash := "delete-test-hash"
	ctx := context.WithValue(context.Background(), config.CountryCode, "FR")

	saveReq := webModels.SaveRequest{
		Message:           encrypted[24:],
		TransmissionNonce: nonce[:],
		Hash:              testHash,
		PublicKey:         browserPublicKey[:],
		Time:              100,
		CostFactor:        10,
	}

	err = testApp.ProcessSave(ctx, saveReq)
	require.NoError(t, err)

	// Delete the message
	testApp.ProcessDelete(context.Background(), testHash)

	// Try to read - should return nil
	recipientPubKey, _, err := box.GenerateKey(rand.Reader)
	require.NoError(t, err)

	readReq := webModels.ReadRequest{
		Hash:      testHash,
		PublicKey: recipientPubKey[:],
		Password:  false,
	}

	result, _, err := testApp.ProcessRead(context.Background(), readReq)
	assert.NoError(t, err)
	assert.Nil(t, result, "message should be deleted")
}

func TestProcessDelete_WithSpecialChars(t *testing.T) {
	setupTest(t)

	tests := []struct {
		name string
		hash string
	}{
		{
			name: "hash with slash",
			hash: "test/hash/with/slashes",
		},
		{
			name: "hash with query params",
			hash: "test?param=value&other=123",
		},
		{
			name: "hash with special chars",
			hash: "hash+with=special&chars",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic
			testApp.ProcessDelete(context.Background(), tt.hash)
		})
	}
}

func TestProcessDeleteExpired(t *testing.T) {
	setupTest(t)

	// Create messages with different expiration times
	browserPublicKey, browserPrivateKey, err := box.GenerateKey(rand.Reader)
	require.NoError(t, err)

	nonce, err := testKeys.GenerateNonce()
	require.NoError(t, err)

	// Message 1: Already expired
	encrypted1 := box.Seal(nonce[:], []byte("expired message"), &nonce, testKeys.PublicKey, browserPrivateKey)

	// Temporarily change default duration to create expired message
	originalDuration := testApp.Config.DefaultDurationTime
	testApp.Config.DefaultDurationTime = -time.Hour // 1 hour ago

	ctx := context.WithValue(context.Background(), config.CountryCode, "PL")

	saveReq1 := webModels.SaveRequest{
		Message:           encrypted1[24:],
		TransmissionNonce: nonce[:],
		Hash:              "expired-hash-1",
		PublicKey:         browserPublicKey[:],
		Time:              100,
		CostFactor:        10,
	}

	err = testApp.ProcessSave(ctx, saveReq1)
	require.NoError(t, err)

	// Restore duration and create active message
	testApp.Config.DefaultDurationTime = originalDuration

	nonce2, err := testKeys.GenerateNonce()
	require.NoError(t, err)

	encrypted2 := box.Seal(nonce2[:], []byte("active message"), &nonce2, testKeys.PublicKey, browserPrivateKey)

	saveReq2 := webModels.SaveRequest{
		Message:           encrypted2[24:],
		TransmissionNonce: nonce2[:],
		Hash:              "active-hash-1",
		PublicKey:         browserPublicKey[:],
		Time:              100,
		CostFactor:        10,
	}

	err = testApp.ProcessSave(ctx, saveReq2)
	require.NoError(t, err)

	// Delete expired messages
	err = testApp.ProcessDeleteExpired(context.Background())
	assert.NoError(t, err)

	// Verify expired message is gone
	recipientPubKey, _, err := box.GenerateKey(rand.Reader)
	require.NoError(t, err)

	readReq1 := webModels.ReadRequest{
		Hash:      "expired-hash-1",
		PublicKey: recipientPubKey[:],
		Password:  false,
	}

	result1, _, err := testApp.ProcessRead(context.Background(), readReq1)
	assert.NoError(t, err)
	assert.Nil(t, result1, "expired message should be deleted")

	// Verify active message still exists
	recipientPubKey2, _, err := box.GenerateKey(rand.Reader)
	require.NoError(t, err)

	readReq2 := webModels.ReadRequest{
		Hash:      "active-hash-1",
		PublicKey: recipientPubKey2[:],
		Password:  false,
	}

	result2, _, err := testApp.ProcessRead(context.Background(), readReq2)
	assert.NoError(t, err)
	assert.NotNil(t, result2, "active message should still exist")
}

func TestProcessSave_VerifyExpiration(t *testing.T) {
	setupTest(t)

	browserPublicKey, _, err := box.GenerateKey(rand.Reader)
	require.NoError(t, err)

	ctx := context.WithValue(context.Background(), config.CountryCode, "DE")

	saveReq := webModels.SaveRequest{
		Message:           []byte("test"),
		TransmissionNonce: make([]byte, 24),
		Hash:              "expiration-test",
		PublicKey:         browserPublicKey[:],
		Time:              100,
		CostFactor:        10,
	}

	beforeSave := time.Now()
	err = testApp.ProcessSave(ctx, saveReq)
	require.NoError(t, err)
	afterSave := time.Now()

	// Verify expiration time is approximately DefaultDurationTime in the future
	expectedExpiration := beforeSave.Add(testConf.DefaultDurationTime)

	// The actual expiration should be within a reasonable range
	assert.True(t,
		expectedExpiration.Before(afterSave.Add(testConf.DefaultDurationTime).Add(time.Second)),
		"expiration time should be approximately DefaultDurationTime in the future")
}
