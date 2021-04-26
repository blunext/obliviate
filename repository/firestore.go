package repository

import (
	"cloud.google.com/go/firestore"
	"context"
	firebase "firebase.google.com/go"
	"fmt"
	"github.com/sirupsen/logrus"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"obliviate/repository/model"
	"time"
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

func NewConnection(ctx context.Context, firestoreCredentialFile string, projectID string, prodEnv bool) *db {

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
		logrus.Errorf("Cannot create new App while connecting to firestore: %v", err)
	}
	client, err := app.Firestore(ctx)
	if err != nil {
		logrus.Errorf("Cannot create new client while connecting to firestore: %v", err)
	}

	d := db{
		messageCollection: collection{coll: "messages", keyColl: "commons", keyDoc: "keys"},
		counter:           Counter{counterShards, "stats", client},
		client:            client,
	}
	logrus.Info("Firestore connected")

	if !d.counter.counterExists(ctx) {
		d.counter.initCounter(ctx)
		logrus.Info("Counter initialized")
	}

	i, _ := d.counter.getCount(ctx)
	logrus.Infof("Counter = %d", i)

	return &d
}

func (d *db) SaveMessage(ctx context.Context, data model.MessageModel) error {
	_, err := d.client.Collection(d.messageCollection.coll).Doc(data.Key()).Set(ctx, data.Message)
	if err != nil {
		return fmt.Errorf("error while saving key: %s, err: %v", data.Key(), err)
	}
	logrus.Debugf("massage saved, key: %v, t: %d, len: %d", data.Key(), data.Message.Time, len(data.Message.Txt))
	return nil
}

func (d *db) GetMessage(ctx context.Context, key string) (model.MessageType, error) {

	data := model.MessageModel{}

	doc, err := d.client.Collection(d.messageCollection.coll).Doc(key).Get(ctx)
	if err != nil {
		if status.Code(err) != codes.NotFound {
			return data.Message, fmt.Errorf("error while getting message, err: %v", err)
		}
		logrus.Trace("message not found")
		return data.Message, nil
	}

	if err := doc.DataTo(&data.Message); err != nil {
		logrus.Trace("message found")
		return data.Message, fmt.Errorf("error mapping data into message struct: %v", err)
	}

	if data.Message.ValidTo.Before(time.Now()) {
		logrus.Warn("massage found but not valid")
		return data.Message, nil
	}

	return data.Message, nil
}

func (d *db) DeleteMessage(ctx context.Context, key string) {
	_, err := d.client.Collection(d.messageCollection.coll).Doc(key).Delete(ctx)
	if err != nil {
		logrus.Errorf("cannot remove doc, key: %v\n", key)
	}
}

func (d *db) DeleteBeforeNow(ctx context.Context) error {

	iter := d.client.Collection(d.messageCollection.coll).Where("valid", "<", time.Now()).Documents(ctx)
	numDeleted := 0
	batch := d.client.Batch()
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to iterate: %v", err)
		}
		batch.Delete(doc.Ref)
		numDeleted++
	}
	// If there are no documents to delete,
	// the process is over.
	if numDeleted == 0 {
		logrus.Trace("Nothing to delete")
		return nil
	}
	_, err := batch.Commit(ctx)
	return err
}

// -----------------------------------------------------------

func (d *db) SaveEncryptedKeys(ctx context.Context, encrypted []byte) error {
	keys := model.Key{Key: encrypted}
	_, err := d.client.Collection(d.messageCollection.keyColl).Doc(d.messageCollection.keyDoc).Set(ctx, keys)
	if err != nil {
		return fmt.Errorf("error while saving encrypted keys: %s, err: %v", keys.Key, err)
	}
	logrus.Info("encrypted keys saved")
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
		return nil, fmt.Errorf("error mapping data into key struct : %v\n", err)
	}
	logrus.Info("encrypted keys fetched from db")
	return key.Key, nil
}

func (d *db) IncreaseCounter(ctx context.Context) {
	_, err := d.counter.incrementCounter(ctx)
	if err != nil {
		logrus.Warningf("Increase counter error: %v", err)
	}
}
