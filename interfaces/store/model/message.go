package model

import (
	"time"
)

type MessageType struct {
	Txt       []byte    `firestore:"txt"`
	ValidTo   time.Time `firestore:"valid"`
	Nonce     []byte    `firestore:"nonce"`
	PublicKey []byte    `firestore:"publicKey"`
}

type MessageModel struct {
	key     string
	Message MessageType
}

func NewMessage(key string, txt []byte, valid time.Time, nonce []byte, publicKey []byte) MessageModel {
	m := MessageModel{
		key: key,
		Message: MessageType{
			Txt:       txt,
			ValidTo:   valid,
			Nonce:     nonce,
			PublicKey: publicKey,
		},
	}
	return m
}

func (m MessageModel) Key() string {
	return m.key
}
