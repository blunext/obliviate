package i18n

import (
	"cmp"
	"context"
	"fmt"
	"html"
	"log/slog"
	"slices"
	"strings"
	"sync"

	"obliviate/logs"

	"golang.org/x/text/language"
)

const (
	// real browsers send far less; parsing cost grows with the length of this untrusted header
	maxAcceptLanguageLength = 1024

	githubLink        = `<a href="https://github.com/blunext/obliviate" target="_blank" rel="noopener noreferrer">GitHub</a>`
	contactLinkFormat = `<a href="mailto:info@securenote.io" target="_blank" rel="noopener noreferrer">%s</a>`
)

// languageFallbacks are appended after the user's languages, for those the matcher maps to English or a distant language.
// Keys are a base language, optionally with a script.
var languageFallbacks = map[string]language.Tag{
	"az":      language.Turkish,
	"rue":     language.Ukrainian,
	"yue":     language.SimplifiedChinese,
	"bs-Cyrl": language.Serbian,
}

type translation map[string]string

type i18n struct {
	translations map[string]translation
	matcher      language.Matcher
	sync.Mutex
}

func NewTranslation() *i18n {

	// English is the default; Russian wins ties between Cyrillic languages, e.g. for Bashkir or Chechen
	languages := []language.Tag{language.English, language.Russian}

	for tag := range translationsSet {
		if tag != language.English && tag != language.Russian {
			languages = append(languages, tag)
		}
	}
	// the matcher resolves ties by list order, so it must not depend on map iteration
	slices.SortFunc(languages[2:], func(a, b language.Tag) int {
		return cmp.Compare(a.String(), b.String())
	})
	tr := i18n{
		matcher:      language.NewMatcher(languages),
		translations: make(map[string]translation),
	}
	return &tr
}

func (t *i18n) GetTranslation(ctx context.Context, acceptLanguage string) translation {
	var acceptedTag language.Tag

	if len(acceptLanguage) > maxAcceptLanguageLength {
		acceptLanguage = ""
	}

	acceptTagList, _, err := language.ParseAcceptLanguage(acceptLanguage)
	if err != nil {
		acceptedTag = language.English
	} else {
		acceptedTag, _, _ = t.matcher.Match(withFallbacks(acceptTagList)...)
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

func withFallbacks(tags []language.Tag) []language.Tag {
	for _, tag := range tags {
		base, _ := tag.Base()
		script, _ := tag.Script()
		if fallback, ok := languageFallbacks[base.String()+"-"+script.String()]; ok {
			tags = append(tags, fallback)
		} else if fallback, ok := languageFallbacks[base.String()]; ok {
			tags = append(tags, fallback)
		}
	}
	return tags
}

// renderInfo escapes the info text and replaces {github} and {contact} placeholders with HTML links
func renderInfo(info, contactLabel string) string {
	r := strings.NewReplacer(
		"{github}", githubLink,
		"{contact}", fmt.Sprintf(contactLinkFormat, html.EscapeString(contactLabel)),
	)
	return r.Replace(html.EscapeString(info))
}
