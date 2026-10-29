package i18n

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testFields struct {
	tag, expected string
}

var testData = []testFields{
	{"pl-PL,pl;q=0.9,en-US;q=0.8,en;q=0.7", "Prywatne, bezpieczne wiadomości"},
	{"fr-CH, fr;q=0.9, en;q=0.8, de;q=0.7, *;q=0.5", "Notes privées sécurisées"},
	{"en-ca,en;q=0.8,en-us;q=0.6,de-de;q=0.4,de;q=0.2", "Private secure notes"},
	{"da, en-gb:q=0.5, en:q=0.4", "Private secure notes"},
}

func TestI18n_GetLazyTranslation(t *testing.T) {

	trans := NewTranslation()
	for _, list := range testData {
		translation := trans.GetTranslation(context.Background(), list.tag)

		var msg string
		var ok bool
		if msg, ok = translation["header"]; !ok {
			assert.True(t, ok, "header not found")
		}
		assert.Equal(t, msg, list.expected, "translation error")
	}

}

func TestI18n_InfoIsRenderedWithLinks(t *testing.T) {
	trans := NewTranslation()
	translation := trans.GetTranslation(context.Background(), "pl")

	assert.Contains(t, translation["info"], "w serwisie "+githubLink+". Jeśli")
	assert.Contains(t, translation["info"], `wyślij nam <a href="mailto:info@securenote.io" target="_blank" rel="noopener noreferrer">wiadomość</a>.`)
}

func TestRenderInfo_EscapesTranslation(t *testing.T) {
	got := renderInfo(`a <b> & {github}'s {contact}`, "<x>")
	assert.Equal(t, `a &lt;b&gt; &amp; `+githubLink+`&#39;s `+
		`<a href="mailto:info@securenote.io" target="_blank" rel="noopener noreferrer">&lt;x&gt;</a>`, got)
}
