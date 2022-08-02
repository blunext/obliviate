package app

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/sirupsen/logrus"

	"obliviate/config"
	"obliviate/crypt"
	"obliviate/handler/webModels"
	"obliviate/repository"
	"obliviate/repository/model"
)

type App struct {
	config *config.Configuration
	keys   *crypt.Keys
	db     repository.DataBase
}

func NewApp(db repository.DataBase, config *config.Configuration, keys *crypt.Keys) *App {
	app := App{
		config: config,
		keys:   keys,
		db:     db,
	}
	return &app
}

func (s *App) ProcessSave(ctx context.Context, request webModels.SaveRequest, country string) error {

	hashEncoded := url.PathEscape(request.Hash)
	messageDataModel := model.NewMessage(hashEncoded, request.Message, time.Now().Add(s.config.DefaultDurationTime),
		request.TransmissionNonce, request.PublicKey, request.Time, request.CostFactor, country)

	err := s.db.SaveMessage(ctx, messageDataModel)
	if err != nil {
		return fmt.Errorf("cannot save message, err: %v", err)
	}

	go func() {
		ct, _ := context.WithTimeout(context.Background(), 3*time.Minute)
		s.db.IncreaseCounter(ct)
	}()

	return nil
}

func (s *App) ProcessRead(ctx context.Context, request webModels.ReadRequest) ([]byte, int, error) {

	hashEncoded := url.PathEscape(request.Hash)

	data, err := s.db.GetMessage(ctx, hashEncoded)
	if err != nil {
		return nil, 0, fmt.Errorf("errod in GetMessage, err: %v", err)
	}
	if data.Txt == nil {
		return nil, 0, nil
	}
	var senderPublicKey [32]byte
	copy(senderPublicKey[:], data.PublicKey)

	var senderNonce [24]byte
	copy(senderNonce[:], data.Nonce)

	decrypted, err := s.keys.BoxOpen(data.Txt, &senderPublicKey, &senderNonce)
	if err != nil {
		return nil, 0, fmt.Errorf("cannot open box, err: %v", err)
	}

	var recipientPublicKey [32]byte
	copy(recipientPublicKey[:], request.PublicKey)

	encrypted, err := s.keys.BoxSeal(decrypted, &recipientPublicKey)
	if err != nil {
		return nil, 0, fmt.Errorf("cannot seal message, err: %v", err)
	}

	if !request.Password {
		// delete only when password is not required
		if s.config.ProdEnv {
			go func() {
				ct, _ := context.WithTimeout(context.Background(), 3*time.Minute)
				s.db.DeleteMessage(ct, hashEncoded)
			}()
		} else { // for testing purposes
			s.db.DeleteMessage(ctx, hashEncoded)
		}
	}

	return encrypted, data.CostFactor, nil
}

func (s *App) ProcessDelete(ctx context.Context, hash string) {
	hashEncoded := url.PathEscape(hash)
	s.db.DeleteMessage(ctx, hashEncoded)
}

func (s *App) ProcessDeleteExpired(ctx context.Context) error {
	if err := s.db.DeleteBeforeNow(ctx); err != nil {
		return fmt.Errorf("delete expired error: %v", err)
	}
	logrus.Trace("Delete expired done")
	return nil
}
