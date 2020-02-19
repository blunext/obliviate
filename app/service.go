package app

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"net/url"
	"obliviate/config"
	"obliviate/crypt"
	"obliviate/interfaces/store"
	"obliviate/interfaces/store/model"
	"time"
)

type App struct {
	config *config.Configuration
	keys   *crypt.Keys
	db     store.Connection
}

func NewApp(db store.Connection, config *config.Configuration, keys *crypt.Keys) *App {
	app := App{
		config: config,
		keys:   keys,
		db:     db,
	}
	return &app
}

func (s *App) ProcessSave(ctx context.Context, message []byte, transmissionNonce []byte, hash string, publicKey []byte, t int) error {

	hashEncoded := url.PathEscape(hash)
	data := model.NewMessage(hashEncoded, message, time.Now().Add(s.config.DefaultDurationTime), transmissionNonce, publicKey, t)

	err := s.db.SaveMessage(ctx, data)
	if err != nil {
		return fmt.Errorf("cannot save message, err: %v", err)
	}
	return nil
}

func (s *App) ProcessRead(ctx context.Context, hash string, publicKey []byte) ([]byte, error) {

	hashEncoded := url.PathEscape(hash)

	data, err := s.db.GetMessage(ctx, hashEncoded)
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
			s.db.DeleteMessage(ct, hashEncoded)
		}()
	} else { // for testing purposes
		s.db.DeleteMessage(ctx, hashEncoded)
	}

	return encrypted, nil
}

func (s *App) ProcessDeleteExpired(ctx context.Context) error {
	if err := s.db.DeleteBeforeNow(ctx); err != nil {
		return fmt.Errorf("delete expired error: %v", err)
	}
	logrus.Trace("Delete expired done")
	return nil
}
