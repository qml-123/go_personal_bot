// Start a web server
package main

import (
	"io"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"code.byted.org/qiumingliang.123/personal_bot/cron"
	"code.byted.org/qiumingliang.123/personal_bot/db"
)

func main() {
	f, err := os.OpenFile("log/log.log", os.O_CREATE|os.O_APPEND|os.O_RDWR, os.ModePerm)
	if err != nil {
		return
	}
	defer func() {
		f.Close()
	}()
	// 组合一下即可，os.Stdout代表标准输出流
	multiWriter := io.MultiWriter(os.Stdout, f)
	logrus.SetOutput(multiWriter)

	db.InitMysqlConn()
	cron.InitCronLoop()
	r := gin.Default()

	post := GetPostRouter()
	for k, v := range post {
		r.POST(k, v)
	}
	r.GET("/ping", Ping)
	if err := r.Run(":8089"); err != nil {
		logrus.WithError(err).Errorf("init fail")
	}
}
