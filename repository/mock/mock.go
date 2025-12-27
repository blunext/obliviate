package mock

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"obliviate/logs"
	model "obliviate/repository/model"
)

type MockDB struct {
	mu           sync.RWMutex
	messageStore map[string]model.MessageModel
	encrypted    []byte
}

func StorageMock() *MockDB {
	d := MockDB{}
	d.messageStore = make(map[string]model.MessageModel)
	return &d
}

func (d *MockDB) SaveMessage(ctx context.Context, data model.MessageModel) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.messageStore[data.Key()] = data
	acceptLanguage := ctx.Value("Accept-Language")
	slog.Info("message saved", logs.Key, data.Key(), logs.AcceptedLang, acceptLanguage)
	return nil
}

func (d *MockDB) GetMessage(ctx context.Context, key string) (model.MessageType, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if m, ok := d.messageStore[key]; ok {
		slog.Debug("key found", logs.Key, m.Key())
		return m.Message, nil
	} else {
		slog.Debug("key not found", logs.Key, m.Key())
		return m.Message, nil
	}
}

func (d *MockDB) GetMessageWithReadLimit(ctx context.Context, key string, maxReads int) (model.MessageType, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	m, ok := d.messageStore[key]
	if !ok {
		slog.Debug("key not found", logs.Key, key)
		return model.MessageType{}, nil
	}

	// Check read count limit
	if m.Message.ReadCount >= maxReads {
		delete(d.messageStore, key)
		return model.MessageType{}, nil
	}

	// Update read count
	m.Message.ReadCount++
	d.messageStore[key] = m

	return m.Message, nil
}

func (d *MockDB) DeleteMessage(ctx context.Context, key string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	delete(d.messageStore, key)
}

func (d *MockDB) DeleteBeforeNow(ctx context.Context) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	for _, v := range d.messageStore {
		if v.Message.ValidTo.Before(time.Now()) {
			// Don't call DeleteMessage here to avoid deadlock
			delete(d.messageStore, v.Key())
		}
	}
	return nil
}

func (d *MockDB) SaveEncryptedKeys(ctx context.Context, encrypted []byte) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.encrypted = encrypted
	return nil
}

func (d *MockDB) GetEncryptedKeys(ctx context.Context) ([]byte, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	return d.encrypted, nil
}

func (d *MockDB) IncreaseCounter(ctx context.Context) {
}
