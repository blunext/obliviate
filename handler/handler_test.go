package handler

import (
	"bytes"
	"crypto/rand"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/nacl/box"
	"golang.org/x/crypto/nacl/secretbox"
	"net/http"
	"net/http/httptest"
	"obliviate/app"
	"obliviate/config"
	"obliviate/crypt"
	"obliviate/crypt/rsa"
	"obliviate/repository"
	"obliviate/repository/mock"
	"os"
	"testing"
	"time"
)

type testParams struct {
	status  int
	message string
}

var conf *config.Configuration
var db repository.DataBase

var params = []testParams{

	{http.StatusOK, "wiadomość"},
	{http.StatusOK, "Facebook i Instagram deklarują w swoich regułach, że nie chcą być agencją rekrutacyjną biznesu pornograficznego ani robić za sutenera. " +
		"Zgodnie z wytycznymi więc oferowanie lub szukanie nagich zdjęć, rozmów erotycznych lub po prostu partnera czy partnerki seksualnej przez wymienione platformy jest zakazane. " +
		"Używanie do tego ikon emoji specyficznych dla danego kontekstu i powszechnie uważanych za nacechowane seksualnie jest, jak deklaruje platforma, dużym przewinieniem. " +
		"Na tyle dużym, że może się skończyć nie tylko ostrzeżeniem, ale wręcz blokadą konta. Chodzi tu między innymi o niewinną tylko z pozoru brzoskwinkę, lśniącego bakłażana " +
		"czy życiodajną kroplę wody."},
	{http.StatusOK, "الخطوط الجوية الفرنسية أو إير فرانس من بين أكبر شركات الطيران في العالم. والمقر " +
		"الرئيسي للشركة في باريس، وهي تابعة لشركة الخطوط الجوية الفرنسية - كيه إل إم، وتنظم الخطوط"},
	{http.StatusOK, "에어 프랑스(프랑스어: Air France 에르 프랑스[*])는 에어 프랑스-KLM의 사업부로 KLM을 합병하기 전에는 프랑스의 국책 항공사였으며, 2009년 9월 기준 종업원수는 60,686명이다[1]. " +
		"본사는 파리 시 근교의 샤를 드 골 공항에 있으며 현재는 에어 프랑스-KLM이 쓰고 있다. 2001년 4월부터 2002년 3월까지 4330만명의 승객을 실어 나르고 125억3천만 유로를 벌어들였다. " +
		"에어 프랑스의 자회사 레지오날은 주로 유럽 내에서 제트 비행기와 터보프롭 비행기로 지역 항공 노선을 운항하고 있다."},
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
	//db = repository.NewConnection(context.Background(), "local", conf.FirestoreCredentialFile, os.Getenv("OBLIVIATE_PROJECT_ID"), conf.ProdEnv)
	db = mock.StorageMock()

}

func TestEncodeDecodeMessage(t *testing.T) {
	rsa := rsa.NewMockAlgorithm()
	//rsa := rsa.NewAlgorithm()
	keys, err := crypt.NewKeys(db, conf, rsa)
	if err != nil {
		logrus.Panicf("cannot create key pair, err: %v", err)
	}
	app := app.NewApp(db, conf, keys)

	for _, tab := range params {

		browserPublicKey, _, _ := box.GenerateKey(rand.Reader)
		_, messageSecretKey, _ := box.GenerateKey(rand.Reader)
		messageNonce, _ := keys.GenerateNonce()

		messageWithNonce := secretbox.Seal(messageNonce[:], []byte(tab.message), &messageNonce, messageSecretKey)
		messageWithSecret := append(messageSecretKey[:], messageWithNonce[24:]...) // take it without nonce

		transmissionNonce, _ := keys.GenerateNonce()
		encryptedTransmission := box.Seal(transmissionNonce[:], messageWithSecret, &transmissionNonce, browserPublicKey, keys.PrivateKey)

		saveRequest := SaveRequest{
			Message:           encryptedTransmission[24:], // take it without nonce, will be base64ed on marshal
			TransmissionNonce: transmissionNonce[:],
			Hash:              makeHash(messageNonce),
			PublicKey:         browserPublicKey[:],
		}

		code, _ := makePost(t, jsonFromStruct(saveRequest), Save(app))
		assert.Equal(t, tab.status, code, "response code not expected")

		// read

		browserPublicKey, browserPrivateKey, _ := box.GenerateKey(rand.Reader) // new keys

		readRequest := ReadRequest{
			Hash:      makeHash(messageNonce),
			PublicKey: browserPublicKey[:],
		}

		code, readResponse := makePost(t, jsonFromStruct(readRequest), Read(app))
		assert.Equal(t, tab.status, code, "response code not expected")

		if code != http.StatusOK {
			continue
		}

		data := ReadResponse{}
		err := json.Unmarshal([]byte(readResponse), &data)
		assert.NoError(t, err, "error unmarshal read response")
		if err != nil {
			continue
		}

		encryptedTransmissionWithNonce := data.Message

		copy(transmissionNonce[:], encryptedTransmissionWithNonce)
		decryptedTransmission, ok := box.Open(nil, encryptedTransmissionWithNonce[24:], &transmissionNonce, keys.PublicKey, browserPrivateKey)

		copy(messageSecretKey[:], decryptedTransmission)

		decryptedMessage, ok := secretbox.Open(nil, decryptedTransmission[32:], &messageNonce, messageSecretKey)
		assert.True(t, ok, "error opening secretbox")
		if !ok {
			continue
		}

		assert.Equal(t, string(decryptedMessage), tab.message, "decrypted message is not the same")
	}
}

func makePost(t *testing.T, jsonMessage []byte, handler http.Handler) (int, string) {
	req, err := http.NewRequest("POST", "/post", bytes.NewBuffer(jsonMessage))
	if err != nil {
		t.Fatal(err)
	}
	r := httptest.NewRecorder()
	handler.ServeHTTP(r, req)
	return r.Code, r.Body.String()
}

func makeHash(in [24]byte) string {
	h := sha1.New()
	h.Write(in[:])
	bs := h.Sum(nil)
	return fmt.Sprintf("%x", bs)
}
