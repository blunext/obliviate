package i18n

import "golang.org/x/text/language"

type translationPair struct {
	key string
	msg string
}

var translationsSet = map[language.Tag][]translationPair{
	language.English: {
		{"title", "Private and secure notes - send your secrets safely."},
		{"header", "Private secure notes"},
		{"description", "Highly secure message encryption open source tool."},
		{"enterTextMessage", "Enter text message to be encrypted"},
		{"secureButton", "Secure message"},
		{"copyLink", "Copy link and send it to a friend. Message will be deleted after being read or after 4 weeks when not read."},
		{"copyLinkButton", "Copy link"},
		{"newMessageButton", "New message"},
		{"decodedMessage", "Decoded message"},
		{"messageRead", "Message was already read, deleted or link is corrupted"},
		{"readMessageButton", "Read message"},
		{"infoHeader", "info about"},
		{"info", "This tool was built with care and respect to your privacy. " +
			"Tool uses various method of encryption to ensure maximum privacy. Tool is open source and code is publicly accessible. " +
			"Feel free to look and see how it was made. More info you can find on"},
		{"linkCorrupted", "Link is corrupted"},
		{"generalError", "Something went wrong. Try again later."},
	},
	language.Polish: {
		{"title", "Prywatne bezpieczne wiadomości"},
		{"header", "Prywatne wiadomości"},
		{"description", "Bezpieczene zakodowane wiadomości"},
		{"enterTextMessage", "Wprowadź wiadomość"},
		{"secureButton", "Zaszufruj wiadomość"},
		{"copyLink", "Skopiuj link i prześlij do przyjaciela. Wiadomość będzie skasowana po odczytaniu lub po 4 tygodniach."},
		{"copyLinkButton", "Skopiuj link"},
		{"newMessageButton", "Nowa wiadomość"},
		{"decodedMessage", "Odszyfrowana wiadomość"},
		{"messageRead", "Wiadomość odczytana, przeterminowana lub link jest błędny"},
		{"readMessageButton", "Odszyfruj wiadomość"},
		{"infoHeader", "opis"},
		{"info", "This tool was built with care and respect to your privacy. " +
			"Tool uses various method of encryption to ensure maximum privacy. Tool is open source and code is publicly accessible. " +
			"Feel free to look and see how it was made. More info you can find on"},
		{"linkCorrupted", "Link uszkodzony"},
		{"generalError", "Coś poszło nie tak. Spróbuj ponownie za jakiś czas."},
	},
}
