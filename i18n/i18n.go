package i18n

import (
	"context"
	"log/slog"
	"obliviate/logs"
	"sync"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
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

	for tag, oneLanguage := range translationsSet {
		if tag != language.English {
			languages = append(languages, tag)
		}
		for _, entry := range oneLanguage {
			err := message.SetString(tag, entry.key, entry.msg)
			if err != nil {
				slog.Error("pair population error", logs.Error, err)
			}
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
	printer := message.NewPrinter(acceptedTag)

	for tag, oneLanguage := range translationsSet {
		if tag.String()[:2] == acceptedBaseLang {
			for _, entry := range oneLanguage {
				tran[entry.key] = printer.Sprintf(entry.key)
			}
		}
	}

	if len(tran) == 0 {
		slog.ErrorContext(ctx, "could not determine translation", logs.LanguageTag, acceptedTag, logs.Language, acceptedBaseLang)
	}

	t.translations[acceptedBaseLang] = tran

	slog.InfoContext(ctx, "language created", logs.Language, acceptedBaseLang)
	return tran
}
