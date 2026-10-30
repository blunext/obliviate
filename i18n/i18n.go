package i18n

import (
	"context"
	"fmt"
	"html"
	"log/slog"
	"strings"
	"sync"

	"obliviate/logs"

	"golang.org/x/text/language"
)

const (
	githubLink        = `<a href="https://github.com/blunext/obliviate" target="_blank" rel="noopener noreferrer">GitHub</a>`
	contactLinkFormat = `<a href="mailto:info@securenote.io" target="_blank" rel="noopener noreferrer">%s</a>`
)

type translation map[string]string

type i18n struct {
	translations map[string]translation
	matcher      language.Matcher
	sync.Mutex
}

func NewTranslation() *i18n {

	var languages []language.Tag

	languages = append(languages, language.English)

	for tag := range translationsSet {
		if tag != language.English {
			languages = append(languages, tag)
		}
	}
	tr := i18n{
		matcher:      language.NewMatcher(languages),
		translations: make(map[string]translation),
	}
	return &tr
}

func (t *i18n) GetTranslation(ctx context.Context, acceptLanguage string) translation {
	var acceptedTag language.Tag

	acceptTagList, _, err := language.ParseAcceptLanguage(acceptLanguage)
	if err != nil {
		acceptedTag = language.English
	} else {
		acceptedTag, _, _ = t.matcher.Match(acceptTagList...)
	}
	acceptedBaseLang := acceptedTag.String()[:2]

	t.Lock()
	defer t.Unlock()

	if tran, ok := t.translations[acceptedBaseLang]; ok {
		slog.InfoContext(ctx, "translation exists", logs.Language, acceptedBaseLang)
		return tran
	}

	tran := translation{}

	for tag, oneLanguage := range translationsSet {
		if tag.String()[:2] == acceptedBaseLang {
			for _, entry := range oneLanguage {
				tran[entry.key] = entry.msg
			}
		}
	}

	if len(tran) == 0 {
		slog.ErrorContext(ctx, "could not determine translation", logs.LanguageTag, acceptedTag, logs.Language, acceptedBaseLang)
	} else {
		tran["info"] = renderInfo(tran["info"], tran["contactLink"])
	}

	t.translations[acceptedBaseLang] = tran

	slog.InfoContext(ctx, "language created", logs.Language, acceptedBaseLang)
	return tran
}

// renderInfo escapes the info text and replaces {github} and {contact} placeholders with HTML links
func renderInfo(info, contactLabel string) string {
	r := strings.NewReplacer(
		"{github}", githubLink,
		"{contact}", fmt.Sprintf(contactLinkFormat, html.EscapeString(contactLabel)),
	)
	return r.Replace(html.EscapeString(info))
}
