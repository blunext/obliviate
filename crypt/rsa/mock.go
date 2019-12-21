package rsa

import (
	"obliviate/config"
)

type mock struct {
	plaintext []byte
}

func NewMockAlgorithm() *mock {
	return &mock{}
}

func (m *mock) EncryptRSA(conf *config.Configuration, plaintext []byte) ([]byte, error) {
	m.plaintext = plaintext
	return []byte("38290823908 32908329083290382903890389032890238903829308290382903890283092803829083098" +
		"sdfdsf dsfds jfidso fjdsiofdsuifduifhdsuifhdsu ifdsuifhsui89ys9u uhuifhikhsuishafidskhdo sads " +
		"sdfdsf dsfds jfidso fjdsiofdsuifduifhdsuifhdsu ifdsuifhsui89ys9u uhuifhikhsuishafidskhdo sads " +
		"sdfdsf dsfds jfidso fjdsiofdsuifduifhdsuifhdsu ifdsuifhsui89ys9u uhuifhikhsuishafidskhdo sads " +
		"sdfdsf dsfds jfidso fjdsiofdsuifduifhdsuifhdsu ifdsuifhsui89ys9u uhuifhikhsuishafidskhdo sads " +
		"sdfdsf dsfds jfidso fjdsiofdsuifduifhdsuifhdsu ifdsuifhsui89ys9u uhuifhikhsuishafidskhdo sads " +
		"dskaop dksa0ksaopdisa0-disakdopsajdospdi0s-a dijks0pdjksds-aidk9saodjsadjs9dusaj0oidjhsdsdj9s adjsd90saidj"), nil
}

func (m *mock) DecryptRSA(conf *config.Configuration, ciphertext []byte) ([]byte, error) {
	return m.plaintext, nil
}
