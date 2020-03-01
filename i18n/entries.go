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
			"Tool uses various method of encryption to ensure maximum privacy. To increase security feel free to use password. " +
			"Tool is Open Source and code is publicly accessible. " +
			"Feel free to look and see how it was made. More info you can find on "},
		{"linkCorrupted", "Link is corrupted"},
		{"generalError", "Something went wrong. Try again later."},
		{"encryptNetworkError", "Something went wrong. Cannot save the message. Please try again."},
		{"decryptNetworkError", "Something went wrong. Cannot load the message. Please try again."},
		{"password", "Password"},
		{"passwordEncodePlaceholder", "Use password to increase security"},
		{"passwordDecodePlaceholder", "enter password"},
		{"linkIsCorrupted", "Link is corrupted"},
		{"ieWarning", "Internet Explorer detected. Operation may take few seconds. Please be patient."},
	},
	language.Polish: {
		{"title", "Prywatne bezpieczne wiadomości"},
		{"header", "Prywatne wiadomości"},
		{"description", "Bezpieczne zakodowane wiadomości"},
		{"enterTextMessage", "Wprowadź wiadomość"},
		{"secureButton", "Szyfruj wiadomość"},
		{"copyLink", "Skopiuj link i prześlij do przyjaciela. Wiadomość będzie skasowana natychmiast po odczytaniu lub po 4 tygodniach."},
		{"copyLinkButton", "Skopiuj link"},
		{"newMessageButton", "Nowa wiadomość"},
		{"decodedMessage", "Odszyfrowana wiadomość"},
		{"messageRead", "Wiadomość odczytana, przeterminowana lub link jest błędny"},
		{"readMessageButton", "Odszyfruj wiadomość"},
		{"infoHeader", "opis"},
		{"info", "Narzędzie używa różnych metod szyfrowania, aby zapewnić maksymalne bezpieczeństwo. " +
			"Bez posiadania linku nie ma możliwości odszyfrowania wiadomości. Użyj hasła aby dodatkowo zwiększyć bezpieczeństwo. " +
			"Kod źródłowy narzędzia jest otwarty i możesz go obejrzeć w serwisie "},
		{"linkCorrupted", "Link uszkodzony"},
		{"generalError", "Coś poszło nie tak. Spróbuj ponownie za jakiś czas."},
		{"encryptNetworkError", "Coś poszło nie tak. Nie mogę zapisać wiadomości. Spróbuj ponownie."},
		{"decryptNetworkError", "Coś poszło nie tak. Nie mogę odczytać wiadomości. Spróbuj ponownie."},
		{"password", "Hasło"},
		{"passwordEncodePlaceholder", "Użyj hasła aby zwiększyć bezpieczeństwo"},
		{"passwordDecodePlaceholder", "wprowadź hasło"},
		{"linkIsCorrupted", "Link jest uszkodzony"},
		{"ieWarning", "Używasz Internet Explorera. Operacja może potrwać parę sekund. Proszę o cierpliwość."},
	},
}
