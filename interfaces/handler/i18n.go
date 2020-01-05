package handler

import (
	"github.com/sirupsen/logrus"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

type entry struct {
	tag language.Tag
	key string
	msg string
}

var entries = [...]entry{
	{language.English, "title", "Private and secure notes - send your secrets safely."},
	{language.English, "header", "Private secure notes"},
	{language.English, "description", "Highly secure message encryption open source tool."},
	{language.English, "enterTextMessage", "Enter text message to be encrypted"},
	{language.English, "secureButton", "Secure message"},
	{language.English, "copyLink", "Copy link and send it to a friend. Message will be deleted after being read or after 4 weeks when not read."},
	{language.English, "copyLinkButton", "Copy link"},
	{language.English, "newMessageButton", "New message"},
	{language.English, "decodedMessage", "Decoded message"},
	{language.English, "messageRead", "Message was already read, deleted or link is corrupted"},
	{language.English, "readMessageButton", "Read message"},
	{language.English, "infoHeader", "info about"},
	{language.English, "info", "This tool was built with care and respect to your privacy. " +
		"Tool uses various method of encryption to ensure maximum privacy. Tool is open source and code is publicly accessible. " +
		"Feel free to look and see how it was made. More info you can find on"},
	{language.English, "linkCorrupted", "Link is corrupted"},
	{language.English, "generalError", "Something went wrong. Try again later."},

	{language.Polish, "title", "Prywatne bezpieczne wiadomości"},
	{language.Polish, "header", "Prywatne wiadomości"},
	{language.Polish, "description", "Bezpieczene zakodowane wiadomości"},
	{language.Polish, "enterTextMessage", "Wprowadź wiadomość"},
	{language.Polish, "secureButton", "Zaszufruj wiadomość"},
	{language.Polish, "copyLink", "Skopiuj link i prześlij do przyjaciela. Wiadomość będzie skasowana po odczytaniu lub po 4 tygodniach."},
	{language.Polish, "copyLinkButton", "Skopiuj link"},
	{language.Polish, "newMessageButton", "Nowa wiadomość"},
	{language.Polish, "decodedMessage", "Odszyfrowana wiadomość"},
	{language.Polish, "messageRead", "Wiadomość odczytana, przeterminowana lub link jest błędny"},
	{language.Polish, "readMessageButton", "Odszyfruj wiadomość"},
	{language.Polish, "infoHeader", "opis"},
	{language.Polish, "info", "This tool was built with care and respect to your privacy. " +
		"Tool uses various method of encryption to ensure maximum privacy. Tool is open source and code is publicly accessible. " +
		"Feel free to look and see how it was made. More info you can find on"},
	{language.Polish, "linkCorrupted", "Link uszkodzony"},
	{language.Polish, "generalError", "Coś poszło nie tak. Spróbuj ponownie za jakiś czas."},
}

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

	tr := translations{matcher: language.NewMatcher(langs)}
	tr.list = make(map[language.Tag]*trans)
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

	logrus.Trace("language %v created", tag)
	return &tran, nil
}
