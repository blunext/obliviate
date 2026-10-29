package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"time"

	"obliviate/config"
	"obliviate/crypt"
	"obliviate/handler/webModels"
	"obliviate/repository"
	"obliviate/repository/model"
)

const maxPasswordMessageReads = 10

type App struct {
	Config *config.Configuration
	keys   *crypt.Keys
	db     repository.DataBase
}

func NewApp(db repository.DataBase, config *config.Configuration, keys *crypt.Keys) *App {
	app := App{
		Config: config,
		keys:   keys,
		db:     db,
	}
	return &app
}

func (s *App) ProcessSave(ctx context.Context, request webModels.SaveRequest) error {
	hashEncoded := url.PathEscape(request.Hash)
	countryCode := ctx.Value(config.CountryCode).(string)

	messageDataModel := model.NewMessage(hashEncoded, request.Message, time.Now().Add(s.Config.DefaultDurationTime),
		request.TransmissionNonce, request.PublicKey, request.Time, request.CostFactor, countryCode)

	err := s.db.SaveMessage(ctx, messageDataModel)
	if err != nil {
		return fmt.Errorf("cannot save message, err: %v", err)
	}

	// synchronous: Cloud Run throttles the CPU once the response is sent
	s.db.IncreaseCounter(ctx)

	return nil
}

func (s *App) ProcessRead(ctx context.Context, request webModels.ReadRequest) ([]byte, int, error) {
	hashEncoded := url.PathEscape(request.Hash)

	// the stored data decides whether the message is deleted now or has a read limit (password);
	// it happens in the same transaction, before the message is decrypted
	data, err := s.db.ReadMessage(ctx, hashEncoded, maxPasswordMessageReads)
	if err != nil {
		return nil, 0, fmt.Errorf("error in GetMessage, err: %v", err)
	}
	if data.Txt == nil {
		return nil, 0, nil
	}

	var senderPublicKey [32]byte
	copy(senderPublicKey[:], data.PublicKey)
	defer crypt.ZeroBytes(&senderPublicKey)

	var senderNonce [24]byte
	copy(senderNonce[:], data.Nonce)
	defer crypt.ZeroBytes(&senderNonce)

	decrypted, err := s.keys.BoxOpen(data.Txt, &senderPublicKey, &senderNonce)
	if err != nil {
		return nil, 0, fmt.Errorf("cannot open box, err: %v", err)
	}

	// destroy plain text message after use
	defer crypt.ZeroBytes(&decrypted)

	var recipientPublicKey [32]byte
	copy(recipientPublicKey[:], request.PublicKey)
	defer crypt.ZeroBytes(&recipientPublicKey)

	encrypted, err := s.keys.BoxSeal(decrypted, &recipientPublicKey)
	if err != nil {
		return nil, 0, fmt.Errorf("cannot seal message, err: %v", err)
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
	slog.InfoContext(ctx, "Delete expired, done")
	return nil
}
