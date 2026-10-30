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
	"az":       language.Turkish,
	"rue":      language.Ukrainian,
	"yue":      language.TraditionalChinese,
	"yue-Hans": language.SimplifiedChinese,
	"bs-Cyrl":  language.Serbian,
}

type translation map[string]string

type i18n struct {
	translations map[language.Tag]translation
	languages    []language.Tag
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
		languages:    languages,
		matcher:      language.NewMatcher(languages),
		translations: make(map[language.Tag]translation),
	}
	return &tr
}

func (t *i18n) GetTranslation(ctx context.Context, acceptLanguage string) translation {
	acceptedTag := language.English

	if len(acceptLanguage) > maxAcceptLanguageLength {
		acceptLanguage = ""
	}

	acceptTagList, _, err := language.ParseAcceptLanguage(acceptLanguage)
	if err == nil {
		for i, tag := range acceptTagList {
			acceptTagList[i] = withChineseScript(tag)
		}
		_, index, _ := t.matcher.Match(withFallbacks(acceptTagList)...)
		acceptedTag = t.languages[index]
	}

	t.Lock()
	defer t.Unlock()

	if tran, ok := t.translations[acceptedTag]; ok {
		slog.InfoContext(ctx, "translation exists", logs.Language, acceptedTag)
		return tran
	}

	tran := translation{}

	for _, entry := range translationsSet[acceptedTag] {
		tran[entry.key] = entry.msg
	}

	if len(tran) == 0 {
		slog.ErrorContext(ctx, "could not determine translation", logs.LanguageTag, acceptedTag)
	} else {
		tran["lang"] = acceptedTag.String()
		tran["info"] = renderInfo(tran["info"], tran["contactLink"])
	}

	t.translations[acceptedTag] = tran

	slog.InfoContext(ctx, "language created", logs.Language, acceptedTag)
	return tran
}

// withChineseScript picks the script for a Chinese tag without one. Left to the matcher, a following bare "zh" beats
// zh-HK, and regions such as MY or US get Traditional Chinese.
func withChineseScript(tag language.Tag) language.Tag {
	base, script, region := tag.Raw()
	if base.String() != "zh" || script.String() != "Zzzz" {
		return tag
	}
	switch region.String() {
	case "TW", "HK", "MO":
		return language.TraditionalChinese
	}
	return language.SimplifiedChinese
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
