package mock

import (
	"context"
	"github.com/sirupsen/logrus"
	model "obliviate/repository/model"
	"time"
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
	logrus.Debugf("massage saved, key: %s", data.Key())
	return nil
}

func (d *db) GetMessage(ctx context.Context, key string) (model.MessageType, error) {
	if m, ok := d.messageStore[key]; ok {
		logrus.Debugf("key found: %s", m.Key())
		return m.Message, nil
	} else {
		logrus.Debugf("key not found: %s", m.Key())
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
