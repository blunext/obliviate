package i18n

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/text/language"
)

type testFields struct {
	tag, expected string
}

var testData = []testFields{
	{"pl-PL,pl;q=0.9,en-US;q=0.8,en;q=0.7", "Prywatne, bezpieczne wiadomości"},
	{"fr-CH, fr;q=0.9, en;q=0.8, de;q=0.7, *;q=0.5", "Notes privées sécurisées"},
	{"en-ca,en;q=0.8,en-us;q=0.6,de-de;q=0.4,de;q=0.2", "Private secure notes"},
	{"da, en-gb:q=0.5, en:q=0.4", "Private secure notes"},
	{"nb", "Private, sikre notater"},
	{"nb-NO", "Private, sikre notater"},
	{"no", "Private, sikre notater"},
	{"nn-NO", "Private, sikre notater"},
	{"da", "Private sikre noter"},
	{"zh-TW", "私密安全筆記"},
	{"zh-CN", "私密安全笔记"},
}

func TestI18n_GetLazyTranslation(t *testing.T) {

	trans := NewTranslation()
	for _, list := range testData {
		translation := trans.GetTranslation(context.Background(), list.tag)

		var msg string
		var ok bool
		if msg, ok = translation["header"]; !ok {
			assert.True(t, ok, "header not found")
		}
		assert.Equal(t, msg, list.expected, "translation error")
	}

}

// dialectData lists Accept-Language values and the translation they should end up with.
var dialectData = []struct {
	expected language.Tag
	tags     []string
}{
	{language.English, []string{
		"en", "en-US", "en-GB", "en-CA", "en-AU", "en-NZ", "en-IE", "en-IN", "en-ZA", "en-SG", "en-PH", "en-NG", "EN-us",
		"az-AZ,az;q=0.9,en-US;q=0.8,en;q=0.7", "rue,en;q=0.5", "yue,en;q=0.5",
		// nothing to match, or the header is invalid
		"", "*", "und", "xx", "!!!", "tlh", "x-klingon", "en-gb:q=0.5",
	}},
	{language.German, []string{
		"de", "de-DE", "de-AT", "de-CH", "de-LI", "de-LU", "de-BE", "de-IT", "de_AT", "gsw", "lb", "rm",
		"de-CH,de;q=0.9,en;q=0.8", "gsw,de;q=0.8",
	}},
	{language.Dutch, []string{"nl", "nl-NL", "nl-BE", "nl-SR", "nl-AW", "fy", "af"}},
	{language.Danish, []string{"da", "da-DK", "da-GL", "fo"}},
	{language.MustParse("nb"), []string{"nb", "nb-NO", "nb-SJ", "no", "no-NO", "nn", "nn-NO"}},
	{language.Swedish, []string{"sv", "sv-SE", "sv-FI", "sv-AX"}},
	{language.Finnish, []string{"fi", "fi-FI"}},
	{language.Estonian, []string{"et", "et-EE"}},
	{language.Latvian, []string{"lv", "lv-LV"}},
	{language.Lithuanian, []string{"lt", "lt-LT"}},
	{language.Hungarian, []string{"hu", "hu-HU", "hu-RO", "hu-SK"}},
	{language.Polish, []string{"pl", "pl-PL", "ar,pl;q=0.5", "*;q=0.5,pl;q=0.1", "en;q=0.1,pl;q=0.9", "az,pl;q=0.5"}},
	{language.Czech, []string{"cs", "cs-CZ"}},
	{language.Slovak, []string{"sk", "sk-SK"}},
	{language.Slovenian, []string{"sl", "sl-SI"}},
	{language.Spanish, []string{
		"es", "es-ES", "es-MX", "es-AR", "es-CO", "es-CL", "es-PE", "es-VE", "es-US", "es-419", "es-GQ",
	}},
	{language.Catalan, []string{"ca", "ca-ES", "ca-AD", "ca-FR", "ca-IT", "ca-ES-valencia", "ca,es;q=0.9", "ca-ES,ca;q=0.9,es;q=0.8,en;q=0.7"}},
	{language.Portuguese, []string{"pt", "pt-PT", "pt-BR", "pt-AO", "pt-MZ", "pt-CV", "pt_BR"}},
	{language.Italian, []string{"it", "it-IT", "it-CH", "it-SM", "it-VA"}},
	{language.Romanian, []string{"ro", "ro-RO", "ro-MD", "mo"}},
	{language.French, []string{
		"fr", "fr-FR", "fr-CA", "fr-BE", "fr-CH", "fr-LU", "fr-CI", "fr-SN", "fr-CM", "fr-CD", "fr-MA",
		"oc", "br", "co", "ht", "wo", "ar,fr;q=0.8",
	}},
	{language.Ukrainian, []string{"uk", "uk-UA", "rue", "rue-UA", "rue-SK"}},
	{language.Russian, []string{
		"ru", "ru-RU", "ru-BY", "ru-KZ", "ru-UA", "ru-MD", "ru-LV",
		"be", "be-BY", "kk", "ky", "tg", "mn", "tt", "sah", "ab", "udm", "ba", "ce", "cv", "os", "av", "kv",
		"uz", "uz-UZ", "uz-Latn", "uz-Cyrl", "tk", "tk-TM", "az,ru;q=0.9",
	}},
	{language.Bulgarian, []string{"bg", "bg-BG", "mk", "mk-MK"}},
	{language.Croatian, []string{"hr", "hr-HR", "hr-BA", "bs", "bs-BA"}},
	{language.Albanian, []string{"sq", "sq-AL", "sq-XK", "sq-MK"}},
	{language.Serbian, []string{
		"sr", "sr-RS", "sr-Latn", "sr-Latn-RS", "sr-Cyrl", "sr-Cyrl-RS", "sr-ME", "sr-BA", "sr-XK", "sh",
		"bs-Cyrl", "bs-Cyrl-BA",
	}},
	{language.Greek, []string{"el", "el-GR", "el-CY"}},
	{language.SimplifiedChinese, []string{
		"zh", "zh-CN", "zh-Hans", "zh-Hans-CN", "zh-Hans-HK", "zh-Hans-TW", "zh-SG", "zh-MY", "zh-ID", "zh-DE", "zh-US", "zh-CA", "zh-GB", "zh-AU",
		"cmn", "cmn-Hans", "wuu", "hak", "nan", "gan", "hsn", "yue-Hans", "yue-CN",
		"zh-CN,zh;q=0.9,en;q=0.8",
	}},
	{language.TraditionalChinese, []string{
		"zh-TW", "zh_TW", "ZH-tw", "zh-HK", "zh-MO", "zh-Hant", "zh-Hant-TW", "zh-Hant-HK", "zh-Hant-MO", "zh-hant-tw",
		"cmn-Hant", "cmn-TW", "yue", "yue-HK", "yue-Hant",
		"zh-TW,zh-Hant;q=0.9", "zh-TW,zh;q=0.9,en-US;q=0.8,en;q=0.7", "zh-HK,zh;q=0.9,en;q=0.8", "zh-MO,zh;q=0.9",
	}},
	{language.Indonesian, []string{"id", "id-ID", "ms", "ms-MY", "ms-SG", "ms-BN", "jv", "su"}},
	{language.Filipino, []string{"fil", "fil-PH", "tl", "tl-PH", "fil-PH,fil;q=0.9,en-US;q=0.8,en;q=0.7"}},
	{language.Turkish, []string{"tr", "tr-TR", "tr-CY", "az", "az-AZ", "az-Latn", "ar,az;q=0.5"}},
	{language.Hindi, []string{"hi", "hi-IN", "mr", "bho", "mai", "awa", "sa", "gu"}},
	{language.Bengali, []string{"bn", "bn-BD", "bn-IN", "as", "syl", "rkt"}},
	{language.Japanese, []string{"ja", "ja-JP"}},
	{language.Korean, []string{"ko", "ko-KR", "ko-KP", "ko-CN", "ko-KR,ko;q=0.9,en-US;q=0.8,en;q=0.7", "ko,pl;q=0.5", "ko,az;q=0.5"}},
	{language.Vietnamese, []string{"vi", "vi-VN", "vi-US", "vi-VN,vi;q=0.9,en-US;q=0.8,en;q=0.7"}},
	{language.Thai, []string{"th", "th-TH", "th-TH,th;q=0.9,en-US;q=0.8,en;q=0.7"}},
}

func titleOf(tag language.Tag) string {
	for _, e := range translationsSet[tag] {
		if e.key == "title" {
			return e.msg
		}
	}
	return ""
}

func TestI18n_TitlesIdentifyLanguages(t *testing.T) {
	seen := map[string]language.Tag{}
	for tag := range translationsSet {
		title := titleOf(tag)
		assert.NotEmpty(t, title, tag)
		assert.NotContains(t, seen, title, tag)
		seen[title] = tag
	}
}

func TestI18n_DialectsMatchClosestLanguage(t *testing.T) {
	for _, data := range dialectData {
		for _, tag := range data.tags {
			translation := NewTranslation().GetTranslation(context.Background(), tag)
			assert.Equal(t, titleOf(data.expected), translation["title"], "%q should use %s", tag, data.expected)
		}
	}
}

func TestI18n_OversizedHeaderIsIgnored(t *testing.T) {
	trans := NewTranslation()

	fits := strings.Repeat("pl,", maxAcceptLanguageLength/3)
	assert.Equal(t, titleOf(language.Polish), trans.GetTranslation(context.Background(), fits)["title"])

	tooLong := strings.Repeat("pl,", maxAcceptLanguageLength/3+1)
	assert.Equal(t, titleOf(language.English), trans.GetTranslation(context.Background(), tooLong)["title"])
}

func TestI18n_HardcodedLanguagesHaveTranslations(t *testing.T) {
	assert.NotEmpty(t, translationsSet[language.English])
	assert.NotEmpty(t, translationsSet[language.Russian])
}

// A translation is cached per language, so a tag requested first must not spoil it for the others.
func TestI18n_TagDoesNotPoisonCache(t *testing.T) {
	for _, data := range dialectData {
		for _, tag := range data.tags {
			trans := NewTranslation()
			trans.GetTranslation(context.Background(), tag)

			translation := trans.GetTranslation(context.Background(), data.expected.String())
			assert.Equal(t, titleOf(data.expected), translation["title"], "after %q", tag)
		}
	}
}

func TestI18n_EveryTagGetsCompleteTranslation(t *testing.T) {
	regions := []string{
		"AL", "AO", "AR", "AT", "AU", "BA", "BD", "BE", "BG", "BR", "BY", "CA", "CH", "CL", "CN", "CO", "CY", "CZ", "DE", "DK",
		"EG", "ES", "FI", "FR", "GB", "GR", "HK", "HR", "HU", "ID", "IN", "IT", "JP", "KE", "KR", "KZ", "LI", "LU", "MD", "ME",
		"MO", "MX", "MY", "MZ", "NG", "NL", "NO", "PK", "PL", "PT", "RO", "RS", "RU", "SA", "SE", "SG", "SI", "SK", "TR", "TW",
		"UA", "US", "XK", "ZA",
	}
	scripts := []string{"Arab", "Cyrl", "Deva", "Hans", "Hant", "Latn"}

	var tags []string
	for a := 'a'; a <= 'z'; a++ {
		for b := 'a'; b <= 'z'; b++ {
			base := string(a) + string(b)
			tags = append(tags, base)
			for _, region := range regions {
				tags = append(tags, base+"-"+region)
			}
			for _, script := range scripts {
				tags = append(tags, base+"-"+script)
			}
		}
	}

	var failed []string
	for _, tag := range tags {
		translation := NewTranslation().GetTranslation(context.Background(), tag)
		for _, e := range translationsSet[language.English] {
			if msg := translation[e.key]; msg == "" || msg == e.key {
				failed = append(failed, tag+" "+e.key)
			}
		}
	}
	assert.Empty(t, failed)
}

func TestI18n_LangIsTheServedLanguage(t *testing.T) {
	for _, data := range dialectData {
		for _, tag := range data.tags {
			translation := NewTranslation().GetTranslation(context.Background(), tag)
			assert.Equal(t, data.expected.String(), translation["lang"], "%q", tag)
		}
	}
}

func TestI18n_LanguagesWithCommonPrefixStaySeparate(t *testing.T) {
	trans := NewTranslation()
	for tag, expected := range map[string]language.Tag{
		"zh-TW": language.TraditionalChinese,
		"zh-CN": language.SimplifiedChinese,
		"fil":   language.Filipino,
		"fi":    language.Finnish,
	} {
		assert.Equal(t, titleOf(expected), trans.GetTranslation(context.Background(), tag)["title"], "%q", tag)
	}
}

func TestI18n_MatcherTiesAreDeterministic(t *testing.T) {
	for i := 0; i < 50; i++ {
		for _, tag := range []string{"ba", "ce", "cv", "os"} {
			translation := NewTranslation().GetTranslation(context.Background(), tag)
			assert.Equal(t, "Приватные безопасные заметки", translation["header"], tag)
		}
	}
}

func TestI18n_InfoIsRenderedWithLinks(t *testing.T) {
	trans := NewTranslation()
	translation := trans.GetTranslation(context.Background(), "pl")

	assert.Contains(t, translation["info"], "w serwisie "+githubLink+". Jeśli")
	assert.Contains(t, translation["info"], `wyślij nam <a href="mailto:info@securenote.io" target="_blank" rel="noopener noreferrer">wiadomość</a>.`)
}

func TestI18n_InfoHasPlaceholdersInEveryLanguage(t *testing.T) {
	for tag, entries := range translationsSet {
		var info, contactLink string
		for _, e := range entries {
			switch e.key {
			case "info":
				info = e.msg
			case "contactLink":
				contactLink = e.msg
			}
		}
		assert.Equal(t, 1, strings.Count(info, "{github}"), "%s: {github} count", tag)
		assert.Equal(t, 1, strings.Count(info, "{contact}"), "%s: {contact} count", tag)
		assert.NotEmpty(t, contactLink, "%s: contactLink", tag)
		assert.NotContains(t, renderInfo(info, contactLink), "{", "%s: unreplaced placeholder", tag)
	}
}

func TestRenderInfo_EscapesTranslation(t *testing.T) {
	got := renderInfo(`a <b> & {github}'s {contact}`, "<x>")
	assert.Equal(t, `a &lt;b&gt; &amp; `+githubLink+`&#39;s `+
		`<a href="mailto:info@securenote.io" target="_blank" rel="noopener noreferrer">&lt;x&gt;</a>`, got)
}
