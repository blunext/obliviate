package crypt

import (
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/nacl/box"

	"obliviate/config"
	"obliviate/crypt/rsa"
	"obliviate/repository/mock"
)

var conf *config.Configuration

// var db repository.DataBase

func init() {
	conf = &config.Configuration{
		DefaultDurationTime:     time.Hour * 24 * 7,
		ProdEnv:                 os.Getenv("ENV") == "PROD",
		MasterKey:               os.Getenv("HSM_MASTER_KEY"),
		KmsCredentialFile:       os.Getenv("KMS_CREDENTIAL_FILE"),
		FirestoreCredentialFile: os.Getenv("FIRESTORE_CREDENTIAL_FILE"),
	}
	// conf.Db = repository.Connect(context.Background(), "test")
	// db = mock.StorageMock()
}

func TestKeysGenerationAndStorage(t *testing.T) {
	// Use fresh DB for this test
	testDB := mock.StorageMock()
	rsaAlg := rsa.NewMockAlgorithm()
	// rsa := rsa.NewAlgorithm()

	keys, err := NewKeys(testDB, conf, rsaAlg, true)
	assert.NoError(t, err, "should not be error")

	pubKey := keys.PublicKeyEncoded

	var priv [32]byte
	//nolint:gosimple
	var pub [32]byte
	pub = *keys.PublicKey
	priv = *keys.PrivateKey

	// Re-use the SAME rsaAlg instance because it holds the state (plaintext) needed for decryption
	keys, err = NewKeys(testDB, conf, rsaAlg, true)
	assert.NoError(t, err, "should not be error")

	assert.Equal(t, pubKey, keys.PublicKeyEncoded, "private keys should be the same")
	assert.Equal(t, priv, *keys.PrivateKey, "private keys should be the same")
	assert.Equal(t, pub, *keys.PublicKey, "public keys should be the same")

}

func TestBoxSealAndOpen(t *testing.T) {
	testDB := mock.StorageMock()
	rsaAlg := rsa.NewMockAlgorithm()
	keys, err := NewKeys(testDB, conf, rsaAlg, true)
	require.NoError(t, err)

	senderPublicKey, senderPrivateKey, err := box.GenerateKey(rand.Reader)
	require.NoError(t, err)

	message := []byte("secret message")

	// Encrypt (Seal)
	encrypted, err := keys.BoxSeal(message, senderPublicKey)
	require.NoError(t, err)
	assert.NotNil(t, encrypted)
	assert.True(t, len(encrypted) > 24, "encrypted message should be longer than nonce")

	// Extract nonce (first 24 bytes)
	var decryptNonce [24]byte
	copy(decryptNonce[:], encrypted[:24])
	// encryptedMessage := encrypted[24:]

	// Decrypt (Open) - Note: BoxSeal uses sender's public key as recipient (conceptually reversed in this usage or just how BoxSeal works)
	// Wait, let's check BoxSeal implementation in keys.go:
	// box.Seal(nonce[:], msg, &nonce, recipientPublicKey, keys.PrivateKey)
	// It encrypts FOR recipientPublicKey using keys.PrivateKey (sender).
	// So to open it, we need the recipient's private key and sender's public key.
	// But Keys.BoxOpen implementation:
	// box.Open(nil, encrypted, decryptNonce, senderPublicKey, keys.PrivateKey)
	// This implies Keys.BoxOpen expects to decrypt using keys.PrivateKey.
	// So BoxSeal and BoxOpen in Keys struct seem to be for different directions or roles?

	// Let's re-read keys.go methods.
	// BoxSeal: box.Seal(..., recipientPublicKey, keys.PrivateKey) -> Encrypts from Server (keys) to Recipient.
	// BoxOpen: box.Open(..., senderPublicKey, keys.PrivateKey) -> Decrypts for Server (keys) from Sender.

	// So if I want to test BoxOpen, I need to encrypt something FOR the Server (keys.PublicKey) using some Sender key.
	// And if I want to test BoxSeal, I encrypt FROM Server TO some Recipient, and I should use Recipient's private key to open it.

	// Test 1: BoxSeal (Server -> Recipient)
	recipientPublicKey, recipientPrivateKey, err := box.GenerateKey(rand.Reader)
	require.NoError(t, err)

	encryptedForRecipient, err := keys.BoxSeal(message, recipientPublicKey)
	require.NoError(t, err)

	// Recipient opens it
	var nonce [24]byte
	copy(nonce[:], encryptedForRecipient[:24])
	cipherText := encryptedForRecipient[24:]
	decrypted, ok := box.Open(nil, cipherText, &nonce, keys.PublicKey, recipientPrivateKey)
	require.True(t, ok, "recipient should be able to decrypt")
	assert.Equal(t, message, decrypted)

	// Test 2: BoxOpen (Sender -> Server)
	// Sender encrypts for Server
	var nonce2 [24]byte
	_, err = io.ReadFull(rand.Reader, nonce2[:])
	require.NoError(t, err)

	encryptedForServer := box.Seal(nil, message, &nonce2, keys.PublicKey, senderPrivateKey)
	// BoxOpen expects encrypted message NOT to contain nonce prepended?
	// keys.go: BoxOpen(encrypted []byte, senderPublicKey *[32]byte, decryptNonce *[24]byte)
	// It takes nonce as argument.
	decryptedByServer, err := keys.BoxOpen(encryptedForServer, senderPublicKey, &nonce2)
	require.NoError(t, err)
	assert.Equal(t, message, decryptedByServer)
}

func TestBoxOpen_Errors(t *testing.T) {
	testDB := mock.StorageMock()
	rsaAlg := rsa.NewMockAlgorithm()
	keys, err := NewKeys(testDB, conf, rsaAlg, true)
	require.NoError(t, err)

	senderPublicKey, senderPrivateKey, err := box.GenerateKey(rand.Reader)
	require.NoError(t, err)

	message := []byte("secret message")
	var nonce [24]byte
	_, err = io.ReadFull(rand.Reader, nonce[:])
	require.NoError(t, err)

	encrypted := box.Seal(nil, message, &nonce, keys.PublicKey, senderPrivateKey)

	// 1. Wrong nonce
	var wrongNonce [24]byte
	_, err = keys.BoxOpen(encrypted, senderPublicKey, &wrongNonce)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot make box open")

	// 2. Wrong sender public key
	wrongPub, _, _ := box.GenerateKey(rand.Reader)
	_, err = keys.BoxOpen(encrypted, wrongPub, &nonce)
	assert.Error(t, err)

	// 3. Corrupted message
	corrupted := make([]byte, len(encrypted))
	copy(corrupted, encrypted)
	corrupted[0] ^= 0xFF // Flip bits
	_, err = keys.BoxOpen(corrupted, senderPublicKey, &nonce)
	assert.Error(t, err)
}

func TestGenerateNonce(t *testing.T) {
	testDB := mock.StorageMock()
	rsaAlg := rsa.NewMockAlgorithm()
	keys, err := NewKeys(testDB, conf, rsaAlg, true)
	require.NoError(t, err)

	n1, err := keys.GenerateNonce()
	require.NoError(t, err)
	assert.Len(t, n1, 24)

	n2, err := keys.GenerateNonce()
	require.NoError(t, err)
	assert.Len(t, n2, 24)

	assert.NotEqual(t, n1, n2, "nonces should be different")
	assert.NotEqual(t, [24]byte{}, n1, "nonce should not be zero")
}

// MockDB for error testing
type ErrorMockDB struct {
	mock.MockDB // Embed to inherit methods
	GetErr      error
	SaveErr     error
}

func (m *ErrorMockDB) GetEncryptedKeys(ctx context.Context) ([]byte, error) {
	if m.GetErr != nil {
		return nil, m.GetErr
	}
	return m.MockDB.GetEncryptedKeys(ctx)
}

func (m *ErrorMockDB) SaveEncryptedKeys(ctx context.Context, keys []byte) error {
	if m.SaveErr != nil {
		return m.SaveErr
	}
	return m.MockDB.SaveEncryptedKeys(ctx, keys)
}

func TestNewKeys_Errors(t *testing.T) {
	rsaAlg := rsa.NewMockAlgorithm()

	t.Run("DB Get Error", func(t *testing.T) {
		errDb := &ErrorMockDB{GetErr: fmt.Errorf("db get error")}
		_, err := NewKeys(errDb, conf, rsaAlg, true)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "error retreaving keys")
	})

	t.Run("DB Save Error", func(t *testing.T) {
		// We need Get to return nil (no keys) so it tries to generate and save

		errDb := &ErrorMockDB{
			MockDB:  *mock.StorageMock(),
			SaveErr: fmt.Errorf("db save error"),
		}

		_, err := NewKeys(errDb, conf, rsaAlg, true)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "error storing keys")
	})
}
