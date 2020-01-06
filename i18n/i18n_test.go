package i18n

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type testFields struct {
	tag, expected string
}

var testData = []testFields{
	{"pl-PL,pl;q=0.9,en-US;q=0.8,en;q=0.7", "Prywatne wiadomości"},
	{"fr-CH, fr;q=0.9, en;q=0.8, de;q=0.7, *;q=0.5", "Private secure notes"},
	{"en-ca,en;q=0.8,en-us;q=0.6,de-de;q=0.4,de;q=0.2", "Private secure notes"},
	{"da, en-gb:q=0.5, en:q=0.4", "Private secure notes"},
}

func TestI18n_GetLazyTranslation(t *testing.T) {

	trans := NewTranslation()
	for _, list := range testData {
		translation := trans.GetTranslation(list.tag)

		var msg string
		var ok bool
		if msg, ok = translation["header"]; !ok {
			assert.True(t, ok, "header not found")
		}
		assert.Equal(t, msg, list.expected, "translation error")
	}

}
