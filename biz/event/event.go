package event

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"code.byted.org/qiumingliang.123/personal_bot/biz/model"
	"code.byted.org/qiumingliang.123/personal_bot/conf"
)

func ReceiveEvent(c *gin.Context) {
	var req = &model.ReceiveEventEncrypt{}
	bytes, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		logrus.WithError(err).Errorf("failed to read request")
		return
	}
	err = json.Unmarshal(bytes, req)
	if err != nil {
		logrus.WithError(err).Errorf("unmarshal failed")
		return
	}
	decryptStr, err := Decrypt(req.Encrypt, conf.Conf.EncryptKey)
	if err != nil {
		logrus.WithError(err).Errorf("decrypt error")
		return
	}

	decryptToken := &model.DecryptToken{}
	err = json.Unmarshal([]byte(decryptStr), decryptToken)
	if err != nil {
		logrus.Errorf("Unmarshal failed again")
		return
	}
	if decryptToken.Challenge != "" {
		c.JSON(200, gin.H{
			"challenge": decryptToken.Challenge,
		})
		return
	}

	event := &model.Event{}
	err = json.Unmarshal([]byte(decryptStr), event)
	if err != nil {
		logrus.Errorf("Unmarshal failed, maybe Challenge")
		return
	}
	logrus.Infof("receive event, event: %v", event)
	ctx := context.Background()
	eventType := event.Header.EventType

	switch eventType {
	case "im.message.receive_v1":
		receiveMsgEvent := &model.ReceiveMessageEvent{}
		err = json.Unmarshal([]byte(decryptStr), receiveMsgEvent)
		if err != nil {
			logrus.Errorf("Unmarshal failed, maybe Challenge")
			return
		}
		go func() {
			err = HandleReceiveMessageEvent(ctx, receiveMsgEvent)
			if err != nil {
				logrus.WithError(err).Errorf("handle receive message event failed")
			}
		}()
	default:
		logrus.Info("unhandled event")
	}

	c.JSON(200, gin.H{
		"message": "ok",
	})
	return
}

func Decrypt(encrypt string, key string) (string, error) {
	buf, err := base64.StdEncoding.DecodeString(encrypt)
	if err != nil {
		return "", fmt.Errorf("base64StdEncode Error[%v]", err)
	}
	if len(buf) < aes.BlockSize {
		return "", errors.New("cipher  too short")
	}
	keyBs := sha256.Sum256([]byte(key))
	block, err := aes.NewCipher(keyBs[:sha256.Size])
	if err != nil {
		return "", fmt.Errorf("AESNewCipher Error[%v]", err)
	}
	iv := buf[:aes.BlockSize]
	buf = buf[aes.BlockSize:]
	// CBC mode always works in whole blocks.
	if len(buf)%aes.BlockSize != 0 {
		return "", errors.New("ciphertext is not a multiple of the block size")
	}
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(buf, buf)
	n := strings.Index(string(buf), "{")
	if n == -1 {
		n = 0
	}
	m := strings.LastIndex(string(buf), "}")
	if m == -1 {
		m = len(buf) - 1
	}
	return string(buf[n : m+1]), nil
}
