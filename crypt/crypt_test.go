package crypt

import (
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"obliviate/config"
	"obliviate/crypt/rsa"
	"obliviate/interfaces/store/mock"
	"os"
	"testing"
	"time"
)

var conf *config.Configuration

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

//
//func TestRandomAndChecksum(t *testing.T) {
//
//	texts := []string{
//		"Asseco Resovia Rzeszów już drugi sezon z rzędu zalicza trudny początek rozgrywek PlusLigi. W środę zespół przegrał 1:3 z beniaminkiem PlusLigi, MKS-em Ślepsk Malow Suwałki, czyli drużyną prowadzoną przez byłego szkoleniowca rzeszowian, Andrzeja Kowala.",
//		"基于谷歌Duplex及AI技术实现自然的人与科技互动 用户拨打电话接入人工客服前，智能机器人优先 进行解答，不仅节省了80%人力成本 还提升了200%的工作效率",
//		"Ce nom de domaine a été réservé par l'intermédiaire de Safebrands",
//		//"",
//	}
//
//	pass := []string{
//		"Resovia Rzeszów już drugi sezon z rzędu",
//		"智能机器人优先 进行解答",
//		"réservé par l'intermédiaire de Safebrands",
//		//"",
//	}
//
//	for _, s := range texts {
//		for _, p := range pass {
//			cipherText := Encrypt([]byte(s), p)
//			decrypted, ok := Decrypt(cipherText, p)
//			assert.True(t, ok, "decode failed")
//			assert.Equal(t, []byte(s), decrypted, "encode/decode error")
//		}
//	}
//
//	for len := 10; len < 20; len++ {
//		for sum := 2; sum < 8; sum++ {
//			rand := RandomKey(len, sum)
//			check := Checksum(rand, len, sum)
//			assert.True(t, check, "Checksum not valid")
//		}
//	}
//
//	rand := RandomKey(20, 2)
//	check := Checksum(rand+"a", 20, 2)
//	assert.True(t, !check, "Checksum not valid")
//
//}

func TestKeysGenerationAndStorage(t *testing.T) {

	rsa := rsa.NewMockAlgorithm()
	//rsa := rsa.NewAlgorithm()

	keys, err := NewKeys(conf, rsa)
	assert.NoError(t, err, "should not be error")

	pubKey := keys.PublicKeyEncoded

	var priv [32]byte
	var pub [32]byte
	pub = *keys.PublicKey
	priv = *keys.PrivateKey

	keys, err = NewKeys(conf, rsa)
	assert.NoError(t, err, "should not be error")

	assert.Equal(t, pubKey, keys.PublicKeyEncoded, "private keys should be the same")
	assert.Equal(t, priv, *keys.PrivateKey, "private keys should be the same")
	assert.Equal(t, pub, *keys.PublicKey, "public keys should be the same")

}
