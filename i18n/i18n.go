package i18n

import (
	"sync"

	"github.com/sirupsen/logrus"
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
				logrus.Errorf("pair population error: %v", err)
			}
		}
	}
	tr := i18n{
		matcher:      language.NewMatcher(languages),
		translations: make(map[string]translation),
	}
	return &tr
}

func (t *i18n) GetTranslation(acceptLanguage string) translation {
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
		logrus.Tracef("translation %v exists", acceptedBaseLang)
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
		logrus.Errorf("could not determine translation for acceptedTag = %v, acceptedBaseLang = %v ", acceptedTag, acceptedBaseLang)
	}

	t.translations[acceptedBaseLang] = tran

	logrus.Infof("language created: %v", acceptedBaseLang)
	return tran
}
