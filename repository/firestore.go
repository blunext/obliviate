package repository

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"obliviate/config"
	"obliviate/logs"
	"obliviate/repository/model"
)

const counterShards = 5

type DataBase interface {
	SaveMessage(ctx context.Context, data model.MessageModel) error
	GetMessage(context.Context, string) (model.MessageType, error)
	DeleteMessage(context.Context, string)
	DeleteBeforeNow(context.Context) error
	SaveEncryptedKeys(context.Context, []byte) error
	GetEncryptedKeys(context.Context) ([]byte, error)
	IncreaseCounter(context.Context)
}

type collection struct {
	coll    string
	keyColl string
	keyDoc  string
}

type db struct {
	client            *firestore.Client
	messageCollection collection
	counter           Counter
}

func NewConnection(ctx context.Context, firestoreCredentialFile, projectID, prefix string, prodEnv bool) *db {
	var err error
	var app *firebase.App

	if prodEnv {
		conf := &firebase.Config{ProjectID: projectID}
		app, err = firebase.NewApp(ctx, conf)
	} else {
		sa := option.WithCredentialsFile(firestoreCredentialFile)
		app, err = firebase.NewApp(ctx, nil, sa)
	}

	if err != nil {
		slog.ErrorContext(ctx, "Cannot create new App while connecting to firestore", logs.Error, err)
	}
	client, err := app.Firestore(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "Cannot create new client while connecting to firestore", logs.Error, err)
	}

	d := db{
		messageCollection: collection{coll: prefix + "messages", keyColl: prefix + "commons", keyDoc: "keys"},
		counter:           Counter{counterShards, prefix + "stats", client},
		client:            client,
	}
	slog.InfoContext(ctx, "Firestore connected")

	if !d.counter.counterExists(ctx) {
		if err := d.counter.initCounter(ctx); err != nil {
			slog.ErrorContext(ctx, "Could not initialize the counters", logs.Error, err)
			panic("Could not initialize the counters")
		}
		slog.InfoContext(ctx, "Counter initialized")
	}

	i, _ := d.counter.getCount(ctx)
	slog.InfoContext(ctx, fmt.Sprintf("Counter = %d", i), logs.Counter, i)

	return &d
}

func (d *db) SaveMessage(ctx context.Context, data model.MessageModel) error {
	_, err := d.client.Collection(d.messageCollection.coll).Doc(data.Key()).Set(ctx, data.Message)
	if err != nil {
		return fmt.Errorf("error while saving key: %s, err: %v", data.Key(), err)
	}
	slog.InfoContext(ctx,
		fmt.Sprintf("message saved t: %d, len: %d, c: %s", data.Message.Time, len(data.Message.Txt), data.Message.Country),
		logs.Length, len(data.Message.Txt), logs.Country, data.Message.Country, logs.Time, data.Message.Time, logs.AcceptedLang, ctx.Value(config.AcceptLanguage))
	return nil
}

func (d *db) GetMessage(ctx context.Context, key string) (model.MessageType, error) {

	data := model.MessageModel{}

	doc, err := d.client.Collection(d.messageCollection.coll).Doc(key).Get(ctx)
	if err != nil {
		if status.Code(err) != codes.NotFound {
			return data.Message, fmt.Errorf("error while getting message, err: %v", err)
		}
		slog.InfoContext(ctx, "message not found")
		return data.Message, nil
	}

	if err := doc.DataTo(&data.Message); err != nil {
		slog.InfoContext(ctx, "message found")
		return data.Message, fmt.Errorf("error mapping data into message struct: %v", err)
	}

	if data.Message.ValidTo.Before(time.Now()) {
		slog.WarnContext(ctx, "message found but not valid")
		return data.Message, nil
	}

	return data.Message, nil
}

func (d *db) DeleteMessage(ctx context.Context, key string) {
	_, err := d.client.Collection(d.messageCollection.coll).Doc(key).Delete(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "cannot remove doc", logs.Key, key)
	}
}

func (d *db) DeleteBeforeNow(ctx context.Context) error {
	// https://firebase.google.com/docs/firestore/manage-data/delete-data#go
	numDeleted := 0
	iter := d.client.Collection(d.messageCollection.coll).Where("valid", "<", time.Now()).Documents(ctx)
	bulkWriter := d.client.BulkWriter(ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to iterate: %v", err)
		}
		_, _ = bulkWriter.Delete(doc.Ref)
		numDeleted++
	}
	if numDeleted == 0 {
		slog.WarnContext(ctx, "Nothing to delete")
		bulkWriter.End()
		return nil
	}
	bulkWriter.Flush()
	slog.InfoContext(ctx, fmt.Sprintf("Deleted %d documents", numDeleted), logs.NumDeleted, numDeleted)
	return nil
}

// -----------------------------------------------------------

func (d *db) SaveEncryptedKeys(ctx context.Context, encrypted []byte) error {
	keys := model.Key{Key: encrypted}
	_, err := d.client.Collection(d.messageCollection.keyColl).Doc(d.messageCollection.keyDoc).Set(ctx, keys)
	if err != nil {
		return fmt.Errorf("error while saving encrypted keys: %s, err: %v", keys.Key, err)
	}
	slog.InfoContext(ctx, "encrypted keys saved")
	return nil
}

func (d *db) GetEncryptedKeys(ctx context.Context) ([]byte, error) {
	doc, err := d.client.Collection(d.messageCollection.keyColl).Doc(d.messageCollection.keyDoc).Get(ctx)
	if err != nil {
		if status.Code(err) != codes.NotFound {
			return nil, fmt.Errorf("error while getting encrypted keys: %v", err)
		}
		return nil, nil
	}
	key := model.Key{}
	if err := doc.DataTo(&key); err != nil {
		return nil, fmt.Errorf("error mapping data into key struct: %v\n", err)
	}
	slog.DebugContext(ctx, "encrypted keys fetched from db")
	return key.Key, nil
}

func (d *db) IncreaseCounter(ctx context.Context) {
	_, err := d.counter.incrementCounter(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "Increase counter error", logs.Error, err)
	}
}
