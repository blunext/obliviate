package i18n

import (
	"context"
	"strings"
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
	{"nb", "Private, sikre notater"},
	{"nb-NO", "Private, sikre notater"},
	{"no", "Private, sikre notater"},
	{"nn-NO", "Private, sikre notater"},
	{"da", "Private sikre noter"},
	// zh-TW alone must not put raw keys in the cache shared with zh-CN
	{"zh-TW", "私密安全笔记"},
	{"zh-CN", "私密安全笔记"},
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

func TestI18n_MatcherTiesAreDeterministic(t *testing.T) {
	for i := 0; i < 50; i++ {
		for _, tag := range []string{"ba", "ce", "cv", "os"} {
			translation := NewTranslation().GetTranslation(context.Background(), tag)
			assert.Equal(t, "Приватные безопасные заметки", translation["header"], tag)
		}
	}
}

func TestI18n_InfoIsRenderedWithLinks(t *testing.T) {
	trans := NewTranslation()
	translation := trans.GetTranslation(context.Background(), "pl")

	assert.Contains(t, translation["info"], "w serwisie "+githubLink+". Jeśli")
	assert.Contains(t, translation["info"], `wyślij nam <a href="mailto:info@securenote.io" target="_blank" rel="noopener noreferrer">wiadomość</a>.`)
}

func TestI18n_InfoHasPlaceholdersInEveryLanguage(t *testing.T) {
	for tag, entries := range translationsSet {
		var info, contactLink string
		for _, e := range entries {
			switch e.key {
			case "info":
				info = e.msg
			case "contactLink":
				contactLink = e.msg
			}
		}
		assert.Equal(t, 1, strings.Count(info, "{github}"), "%s: {github} count", tag)
		assert.Equal(t, 1, strings.Count(info, "{contact}"), "%s: {contact} count", tag)
		assert.NotEmpty(t, contactLink, "%s: contactLink", tag)
		assert.NotContains(t, renderInfo(info, contactLink), "{", "%s: unreplaced placeholder", tag)
	}
}

func TestRenderInfo_EscapesTranslation(t *testing.T) {
	got := renderInfo(`a <b> & {github}'s {contact}`, "<x>")
	assert.Equal(t, `a &lt;b&gt; &amp; `+githubLink+`&#39;s `+
		`<a href="mailto:info@securenote.io" target="_blank" rel="noopener noreferrer">&lt;x&gt;</a>`, got)
}
