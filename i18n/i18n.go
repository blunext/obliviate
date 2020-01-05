package i18n

import (
	"github.com/sirupsen/logrus"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

type trans map[string]string

type translations struct {
	list    map[language.Tag]*trans
	matcher language.Matcher
}

func NewTranslation() *translations {

	langsMap := make(map[language.Tag]bool)
	langs := []language.Tag{}

	langs = append(langs, language.English)
	langsMap[language.English] = true

	for _, e := range entries {
		err := message.SetString(e.tag, e.key, e.msg)
		if err != nil {
			logrus.Errorf("translation population error: %v", err)
		}
		if !langsMap[e.tag] {
			langsMap[e.tag] = true
			langs = append(langs, e.tag)
		}
	}

	tr := translations{matcher: language.NewMatcher(langs), list: make(map[language.Tag]*trans)}
	return &tr
}

func (t *translations) GetLazyTranslation(acceptLanguage string, publicKey string) (*trans, error) {

	acceptTagList, _, _ := language.ParseAcceptLanguage(acceptLanguage) //todo: error
	tag, _, _ := t.matcher.Match(acceptTagList...)

	if tran, ok := t.list[tag]; ok {
		logrus.Trace("translation %v exists", tag)
		return tran, nil
	}

	tran := trans{}
	printer := message.NewPrinter(tag)
	acceptedTagBase, _ := tag.Base()

	for _, e := range entries {
		base, _ := e.tag.Base()
		if base == acceptedTagBase {
			tran[e.key] = printer.Sprintf(e.key)
		}
	}
	if len(tran) == 0 {
		logrus.Warn("language %v not found", tag)
		return t.GetLazyTranslation("en", publicKey)
	}

	tran["PublicKey"] = publicKey
	t.list[tag] = &tran

	logrus.Debugf("language created: %v", tag)
	return &tran, nil
}
