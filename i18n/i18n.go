package i18n

import (
	"github.com/sirupsen/logrus"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"sync"
)

type translations map[string]string

type i18n struct {
	list    map[string]translations
	matcher language.Matcher
	sync.Mutex
}

func NewTranslation() *i18n {

	var langs []language.Tag

	langs = append(langs, language.English)

	for tag, transList := range translationsSet {
		if tag != language.English {
			langs = append(langs, tag)
		}
		for _, tr := range transList {
			err := message.SetString(tag, tr.key, tr.msg)
			if err != nil {
				logrus.Errorf("pair population error: %v", err)
			}
		}
	}
	tr := i18n{matcher: language.NewMatcher(langs), list: make(map[string]translations)}
	return &tr
}

func (t *i18n) GetTranslation(acceptLanguage string) translations {

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

	if tran, ok := t.list[acceptedBaseLang]; ok {
		logrus.Tracef("translation %v exists", acceptedBaseLang)
		return tran
	}

	tran := translations{}
	printer := message.NewPrinter(acceptedTag)

	for tag, transList := range translationsSet {
		if tag.String()[:2] == acceptedBaseLang {
			for _, tr := range transList {
				tran[tr.key] = printer.Sprintf(tr.key)
			}
		}
	}

	if len(tran) == 0 {
		logrus.Errorf("could not determine translation for acceptedTag = %v, acceptedBaseLang = %v ", acceptedTag, acceptedBaseLang)
	}

	t.list[acceptedBaseLang] = tran

	logrus.Debugf("language created: %v", acceptedBaseLang)
	return tran
}
