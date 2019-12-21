package handler

import (
	"github.com/sirupsen/logrus"
	"net/http"
	"obliviate/config"
	"obliviate/interfaces/store/mock"
	"os"
	"testing"
	"time"
)

type testParams struct {
	status           int
	message          string
	messageTestJson  string
	responseTestJson string
	response         string
	password         string
}

var conf *config.Configuration

var params = []testParams{
	//{http.StatusBadRequest, "", `{"message":wiadomość}`, jsonString(messJsonKey, jsonErrMsg), ""},
	//{http.StatusBadRequest, "", ``, jsonString(messJsonKey, jsonErrMsg), ""},
	//{http.StatusBadRequest, "", `ddsadad`, jsonString(messJsonKey, jsonErrMsg), ""},
	//{http.StatusBadRequest, "", `{"message":1}`, jsonString(messJsonKey, jsonErrMsg), ""},
	//{http.StatusBadRequest, "", `{message:"1"}`, jsonString(messJsonKey, jsonErrMsg), ""},
	//{http.StatusBadRequest, "", `{message:""}`, jsonString(messJsonKey, jsonErrMsg), ""},
	//{http.StatusBadRequest, "", `{"message":""}`, jsonString(messJsonKey, messageEmpty), ""},
	{http.StatusOK, "wiadomość", "", "", "", ""},
	{http.StatusOK, "Facebook i Instagram deklarują w swoich regułach, że nie chcą być agencją rekrutacyjną biznesu pornograficznego ani robić za sutenera. " +
		"Zgodnie z wytycznymi więc oferowanie lub szukanie nagich zdjęć, rozmów erotycznych lub po prostu partnera czy partnerki seksualnej przez wymienione platformy jest zakazane. " +
		"Używanie do tego ikon emoji specyficznych dla danego kontekstu i powszechnie uważanych za nacechowane seksualnie jest, jak deklaruje platforma, dużym przewinieniem. " +
		"Na tyle dużym, że może się skończyć nie tylko ostrzeżeniem, ale wręcz blokadą konta. Chodzi tu między innymi o niewinną tylko z pozoru brzoskwinkę, lśniącego bakłażana " +
		"czy życiodajną kroplę wody.", "", "", "", ""},
	{http.StatusOK, "الخطوط الجوية الفرنسية أو إير فرانس من بين أكبر شركات الطيران في العالم. والمقر " +
		"الرئيسي للشركة في باريس، وهي تابعة لشركة الخطوط الجوية الفرنسية - كيه إل إم، وتنظم الخطوط", "", "", "", ""},
	{http.StatusOK, "에어 프랑스(프랑스어: Air France 에르 프랑스[*])는 에어 프랑스-KLM의 사업부로 KLM을 합병하기 전에는 프랑스의 국책 항공사였으며, 2009년 9월 기준 종업원수는 60,686명이다[1]. " +
		"본사는 파리 시 근교의 샤를 드 골 공항에 있으며 현재는 에어 프랑스-KLM이 쓰고 있다. 2001년 4월부터 2002년 3월까지 4330만명의 승객을 실어 나르고 125억3천만 유로를 벌어들였다. " +
		"에어 프랑스의 자회사 레지오날은 주로 유럽 내에서 제트 비행기와 터보프롭 비행기로 지역 항공 노선을 운항하고 있다.", "", "", "", ""},
	// ---- the same with password
	{http.StatusOK, "wiadomość", "", "", "", "u892h kHJKsahjk"},
	{http.StatusOK, "Facebook i Instagram deklarują w swoich regułach, że nie chcą być agencją rekrutacyjną biznesu pornograficznego ani robić za sutenera. " +
		"Zgodnie z wytycznymi więc oferowanie lub szukanie nagich zdjęć, rozmów erotycznych lub po prostu partnera czy partnerki seksualnej przez wymienione platformy jest zakazane. " +
		"Używanie do tego ikon emoji specyficznych dla danego kontekstu i powszechnie uważanych za nacechowane seksualnie jest, jak deklaruje platforma, dużym przewinieniem. " +
		"Na tyle dużym, że może się skończyć nie tylko ostrzeżeniem, ale wręcz blokadą konta. Chodzi tu między innymi o niewinną tylko z pozoru brzoskwinkę, lśniącego bakłażana " +
		"czy życiodajną kroplę wody.", "", "", "", "dsaio j89021u jio"},
	{http.StatusOK, "الخطوط الجوية الفرنسية أو إير فرانس من بين أكبر شركات الطيران في العالم. والمقر " +
		"الرئيسي للشركة في باريس، وهي تابعة لشركة الخطوط الجوية الفرنسية - كيه إل إم، وتنظم الخطوط", "", "", "", "289 hiosahk"},
	{http.StatusOK, "에어 프랑스(프랑스어: Air France 에르 프랑스[*])는 에어 프랑스-KLM의 사업부로 KLM을 합병하기 전에는 프랑스의 국책 항공사였으며, 2009년 9월 기준 종업원수는 60,686명이다[1]. " +
		"본사는 파리 시 근교의 샤를 드 골 공항에 있으며 현재는 에어 프랑스-KLM이 쓰고 있다. 2001년 4월부터 2002년 3월까지 4330만명의 승객을 실어 나르고 125억3천만 유로를 벌어들였다. " +
		"에어 프랑스의 자회사 레지오날은 주로 유럽 내에서 제트 비행기와 터보프롭 비행기로 지역 항공 노선을 운항하고 있다.", "", "", "", "djsiao hio"},
}

func init() {
	formatter := new(logrus.TextFormatter)
	formatter.TimestampFormat = "02-01-2006 15:04:05"
	formatter.FullTimestamp = true
	formatter.ForceColors = true
	logrus.SetFormatter(formatter)
	//logrus.SetReportCaller(true)
	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.FatalLevel)

	conf = &config.Configuration{
		DefaultDurationTime:     time.Hour * 24 * 7,
		ProdEnv:                 os.Getenv("ENV") == "PROD",
		MasterKey:               os.Getenv("HSM_MASTER_KEY"),
		KmsCredentialFile:       os.Getenv("KMS_CREDENTIAL_FILE"),
		FirestoreCredentialFile: os.Getenv("FIRESTORE_CREDENTIAL_FILE"),
	}
	//conf.Db = store.Connect(context.Background(), "test")
	conf.Db = mock.StorageMock()
}

func TestEncodeMessage(t *testing.T) {
	//rsa := rsa.NewMockAlgorithm()
	////rsa := rsa.NewAlgorithm()
	//keys, err := crypt.NewKeys(conf, rsa)
	//if err != nil {
	//	logrus.Panicf("nie mogę utworzyć kluczy, err: ", err)
	//}

	//app := app.NewApp(conf, keys)
	//
	//for i, _ := range params {
	//
	//}
}
