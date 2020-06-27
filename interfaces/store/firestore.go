package store

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
	"obliviate/interfaces/store/model"
	"time"
)

type DataBase interface {
	SaveMessage(ctx context.Context, data model.MessageModel) error
	GetMessage(context.Context, string) (model.MessageType, error)
	DeleteMessage(context.Context, string)
	DeleteBeforeNow(context.Context) error
	SaveEncryptedKeys(context.Context, []byte) error
	GetEncryptedKeys(context.Context) ([]byte, error)
}

type db struct {
	client      *firestore.Client
	messageColl string
	keyColl     string
	keyDoc      string
}

func NewConnection(ctx context.Context, messageColl string, firestoreCredentialFile string, projectID string, prodEnv bool) *db {

	d := db{messageColl: messageColl, keyColl: "commons", keyDoc: "keys"}

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
	d.client = client

	//defer client.Close()
	logrus.Info("Firestore connected")

	return &d
}

func (d *db) SaveMessage(ctx context.Context, data model.MessageModel) error {
	_, err := d.client.Collection(d.messageColl).Doc(data.Key()).Set(ctx, data.Message)
	if err != nil {
		return fmt.Errorf("error while saving key: %s, err: %v", data.Key(), err)
	}
	logrus.Debugf("massage saved, key: %v, t: %d, len: %d", data.Key(), data.Message.Time, len(data.Message.Txt))
	return nil
}

func (d *db) GetMessage(ctx context.Context, key string) (model.MessageType, error) {

	data := model.MessageModel{}

	doc, err := d.client.Collection(d.messageColl).Doc(key).Get(ctx)
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
	_, err := d.client.Collection(d.messageColl).Doc(key).Delete(ctx)
	if err != nil {
		logrus.Errorf("cannot remove doc, key: %v\n", key)
	}
}

func (d *db) DeleteBeforeNow(ctx context.Context) error {

	iter := d.client.Collection(d.messageColl).Where("valid", "<", time.Now()).Documents(ctx)
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
	_, err := d.client.Collection(d.keyColl).Doc(d.keyDoc).Set(ctx, keys)
	if err != nil {
		return fmt.Errorf("error while saving encrypted keys: %s, err: %v", keys.Key, err)
	}
	logrus.Debug("encrypted keys saved")
	return nil
}

func (d *db) GetEncryptedKeys(ctx context.Context) ([]byte, error) {
	doc, err := d.client.Collection(d.keyColl).Doc(d.keyDoc).Get(ctx)
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
	logrus.Debug("encrypted keys fetched from db")
	return key.Key, nil
}
