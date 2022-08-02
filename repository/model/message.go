package model

import (
	"time"
)

type MessageType struct {
	Txt        []byte    `firestore:"txt"`
	ValidTo    time.Time `firestore:"valid"`
	Nonce      []byte    `firestore:"nonce"`
	PublicKey  []byte    `firestore:"publicKey"`
	Time       int       `firestore:"time,omitempty"`
	CostFactor int       `firestore:"costFactor,omitempty"`
	Country    string    `firestore:"country,omitempty"`
}

type MessageModel struct {
	key     string
	Message MessageType
}

func NewMessage(key string, txt []byte, valid time.Time, nonce []byte, publicKey []byte, time int, costFactor int, country string) MessageModel {
	m := MessageModel{
		key: key,
		Message: MessageType{
			Txt:        txt,
			ValidTo:    valid,
			Nonce:      nonce,
			PublicKey:  publicKey,
			Time:       time,
			CostFactor: costFactor,
			Country:    country,
		},
	}
	if time != 0 {
		m.Message.Time = time
	}
	if costFactor != 0 {
		m.Message.CostFactor = costFactor
	}
	return m
}

func (m MessageModel) Key() string {
	return m.key
}
