package i18n

import (
	"github.com/sirupsen/logrus"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"sync"
)

type translations map[string]string

type i18n struct {
	list    map[language.Tag]*translations
	matcher language.Matcher
	sync.Mutex
}

func NewTranslation() *i18n {

	langs := []language.Tag{}

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
	tr := i18n{matcher: language.NewMatcher(langs), list: make(map[language.Tag]*translations)}
	return &tr
}

func (t *i18n) GetLazyTranslation(acceptLanguage string, publicKey string) *translations {

	var acceptedTag language.Tag

	acceptTagList, _, err := language.ParseAcceptLanguage(acceptLanguage)
	if err != nil {
		acceptedTag = language.English
	} else {
		acceptedTag, _, _ = t.matcher.Match(acceptTagList...)
	}

	t.Lock()
	defer t.Unlock()

	if tran, ok := t.list[acceptedTag]; ok {
		logrus.Trace("translation %v exists", acceptedTag)
		return tran
	}

	tran := translations{}
	printer := message.NewPrinter(acceptedTag)
	acceptedTagBase, _ := acceptedTag.Base()

	for tag, transList := range translationsSet {
		base, _ := tag.Base()
		if base == acceptedTagBase {
			for _, tr := range transList {
				tran[tr.key] = printer.Sprintf(tr.key)
			}
		}
	}

	tran["PublicKey"] = publicKey
	t.list[acceptedTag] = &tran

	logrus.Debugf("language created: %v", acceptedTag)
	return &tran
}
