package app

import (
	"context"
	"fmt"
	"net/url"
	"obliviate/config"
	"obliviate/crypt"
	"obliviate/interfaces/store/model"
	"time"
)

type App struct {
	config *config.Configuration
	keys   *crypt.Keys
}

func NewApp(config *config.Configuration, keys *crypt.Keys) *App {
	app := App{
		config: config,
		keys:   keys,
	}
	return &app
}

func (s *App) ProcessSave(ctx context.Context, message []byte, transmissionNonce []byte, hash string, publicKey []byte) error {

	hashEncoded := url.PathEscape(hash)
	data := model.NewMessage(hashEncoded, message, time.Now().Add(s.config.DefaultDurationTime), transmissionNonce, publicKey)

	err := s.config.Db.SaveMessage(ctx, data)
	if err != nil {
		return fmt.Errorf("cannot save message, err: %v", err)
	}
	return nil
}

func (s *App) ProcessRead(ctx context.Context, hash string, publicKey []byte) ([]byte, error) {

	hashEncoded := url.PathEscape(hash)

	data, err := s.config.Db.GetMessage(ctx, hashEncoded)
	if err != nil {
		return nil, fmt.Errorf("errod in GetMessage, err: %v", err)
	}
	if data.Txt == nil {
		return nil, nil
	}
	var senderPublicKey [32]byte
	copy(senderPublicKey[:], data.PublicKey)

	var senderNonce [24]byte
	copy(senderNonce[:], data.Nonce)

	decrypted, err := s.keys.BoxOpen(data.Txt, &senderPublicKey, &senderNonce)
	if err != nil {
		return nil, fmt.Errorf("cannot open box, err: %v", err)
	}

	var recipientPublicKey [32]byte
	copy(recipientPublicKey[:], publicKey)

	encrypted, err := s.keys.BoxSeal(decrypted, &recipientPublicKey)
	if err != nil {
		return nil, fmt.Errorf("cannot seal message, err: %v", err)
	}

	if s.config.ProdEnv {
		go func() {
			ct, _ := context.WithTimeout(context.Background(), 3*time.Minute)
			s.config.Db.DeleteMessage(ct, hashEncoded)
		}()
	} else { // for testing purposes
		s.config.Db.DeleteMessage(ctx, hashEncoded)
	}

	return encrypted, nil
}
