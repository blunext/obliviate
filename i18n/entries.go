package i18n

import "golang.org/x/text/language"

type translationPair struct {
	key string
	msg string
}

var translationsSet = map[language.Tag][]translationPair{
	language.English: {
		{"title", "Private and secure notes - send your secrets safely"},
		{"header", "Private secure notes"},
		{"description", "Highly secure message encryption tool that auto-destruct"},
		{"enterTextMessage", "Enter text message to encrypt"},
		{"secureButton", "Encrypt message"},
		{"copyLink", "Copy link and send it to a friend. Message will be deleted after being read or after 4 weeks when not read."},
		{"copyLinkButton", "Copy link"},
		{"newMessageButton", "New message"},
		{"decodedMessage", "Decrypted message"},
		{"messageRead", "Message was already read, deleted or link is corrupted"},
		{"readMessageButton", "Read message"},
		{"infoHeader", "info about"},
		{"info", "This tool was built with care and respect to your privacy. " +
			"To increase security please use use a password. " +
			"Tool is Open Source and code is publicly accessible. You can check it out on "},
		{"info1", "If you want to contact us please send an"},
		{"info2", "email"},
		{"info3", "All translation initiatives are welcome."},
		{"generalError", "Something went wrong. Try again later."},
		{"encryptNetworkError", "Something went wrong. Cannot save the message. Please try again."},
		{"decryptNetworkError", "Something went wrong. Cannot load the message. Please try again."},
		{"password", "Password"},
		{"passwordEncryptPlaceholder", "Optional password to increase security"},
		{"passwordDecryptPlaceholder", "enter password"},
		{"linkIsCorrupted", "Link is corrupted"},
		{"ieEncryptWarning", "Internet Explorer detected. Encryption may take a little longer. Please be patient."},
		{"ieDecryptWarning", "Internet Explorer detected. Decryption may take a little longer. Please be patient."},
	},
	language.Norwegian: {
		{"title", "Private og sikre notater - send hemmelighetene dine trygt"},
		{"header", "Private, sikre notater"},
		{"description", "Sikkert meldingkrypteringsverktøy som selvdestruerer"},
		{"enterTextMessage", "Tast inn melding for kryptering"},
		{"secureButton", "Krypter melding"},
		{"copyLink", "Kopier lenke og send til en venn. Meldingen vil bli slettet etter å ha blitt lest eller etter 4 uker uten å ha blitt åpnet"},
		{"copyLinkButton", "Kopier lenke"},
		{"newMessageButton", "Ny melding"},
		{"decodedMessage", "Dekrypter melding"},
		{"messageRead", "Melding har allerede blitt lest, slettet eller lenke er korrumpert"},
		{"readMessageButton", "Les melding"},
		{"infoHeader", "Info om"},
		{"info", "Dette verktøyet er laget med hensyn og respekt med tanke på ditt personvern. " +
			"For bedre sikkerhet, vennligst bruk et passord. Verktøyet er open-source og lesbar for offentligheten. " +
			"Koden er tilgjengelig på "},
		{"info1", "Om du ønsker å ta kontakt, vennligst send oss "},
		{"info2", "e-post"},
		{"info3", ""},
		{"generalError", "Noe gikk galt. Vennligst prøv igjen senere"},
		{"encryptNetworkError", "Noe gikk galt, meldingen kunne ikke bli lagret. Vennligst prøv igjen."},
		{"decryptNetworkError", "Noe gikk galt, meldingen kunne ikke bli lastet inn. Vennligst prøv igjen."},
		{"password", "Passord"},
		{"passwordEncryptPlaceholder", "Valgfritt passord for høyere sikkerhet"},
		{"passwordDecryptPlaceholder", "Tast inn passord"},
		{"linkIsCorrupted", "Lenken er korrumpert"},
		{"ieDecryptWarning", "Internet Explorer oppdaget. Kryptering kan ta litt lengre tid. Vennligst vær tålmodig."},
		{"ieEncryptWarning", "Internet Explorer oppdaget. Dekryptering kan ta litt lengre tid. Vennlgst vær tålmodig."},
	},
	language.Polish: {
		{"title", "Prywatne bezpieczne wiadomości"},
		{"header", "Prywatne wiadomości"},
		{"description", "Bezpieczne szyfrowane wiadomości"},
		{"enterTextMessage", "Wpisz wiadomość do zaszyfrowania"},
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
			"Kod źródłowy narzędzia jest otwarty i możesz go obejrzeć w serwisie"},
		{"info1", "Jeśli chcesz się z nami skontaktować wyślij nam"},
		{"info2", "wiadomość"},
		{"info3", ""},
		{"generalError", "Coś poszło nie tak. Spróbuj ponownie za jakiś czas."},
		{"encryptNetworkError", "Coś poszło nie tak. Nie mogę zapisać wiadomości. Spróbuj ponownie."},
		{"decryptNetworkError", "Coś poszło nie tak. Nie mogę odczytać wiadomości. Spróbuj ponownie."},
		{"password", "Hasło"},
		{"passwordEncryptPlaceholder", "Możesz użyć hasła aby zwiększyć bezpieczeństwo"},
		{"passwordDecryptPlaceholder", "wprowadź hasło"},
		{"linkIsCorrupted", "Link jest uszkodzony"},
		{"ieEncryptWarning", "Używasz Internet Explorera. Szyfrowanie może potrwać trochę dłużej. Proszę o cierpliwość."},
		{"ieDecryptWarning", "Używasz Internet Explorera. Odszyfrowanie może potrwać trochę dłużej. Proszę o cierpliwość."},
	},
}
