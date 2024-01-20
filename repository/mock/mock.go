package mock

import (
	"context"
	"log/slog"
	"obliviate/logs"
	"time"

	model "obliviate/repository/model"
)

type db struct {
	messageStore map[string]model.MessageModel
	encrypted    []byte
}

func StorageMock() *db {
	d := db{}
	d.messageStore = make(map[string]model.MessageModel)
	return &d
}

func (d *db) SaveMessage(ctx context.Context, data model.MessageModel) error {
	d.messageStore[data.Key()] = data
	acceptLanguage := ctx.Value("Accept-Language")
	slog.Info("message saved", logs.Key, data.Key(), logs.AcceptedLang, acceptLanguage)
	return nil
}

func (d *db) GetMessage(ctx context.Context, key string) (model.MessageType, error) {
	if m, ok := d.messageStore[key]; ok {
		slog.Debug("key found", logs.Key, m.Key())
		return m.Message, nil
	} else {
		slog.Debug("key not found", logs.Key, m.Key())
		return m.Message, nil
	}
}

func (d *db) DeleteMessage(ctx context.Context, key string) {
	delete(d.messageStore, key)
}

func (d *db) DeleteBeforeNow(ctx context.Context) error {
	for _, v := range d.messageStore {
		if v.Message.ValidTo.Before(time.Now()) {
			d.DeleteMessage(ctx, v.Key())
		}
	}
	return nil
}

func (d *db) SaveEncryptedKeys(ctx context.Context, encrypted []byte) error {
	d.encrypted = encrypted
	return nil
}

func (d *db) GetEncryptedKeys(ctx context.Context) ([]byte, error) {
	return d.encrypted, nil
}

func (d *db) IncreaseCounter(ctx context.Context) {
}
