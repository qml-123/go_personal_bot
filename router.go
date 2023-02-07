package main

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"code.byted.org/qiumingliang.123/personal_bot/biz/event"
)

var post = map[string]gin.HandlerFunc{
	"/webhook/event": event.ReceiveEvent,
}

func GetPostRouter() map[string]gin.HandlerFunc {
	return post
}

func Ping(c *gin.Context) {
	logrus.Info("a sample app log")
	c.JSON(200, gin.H{
		"message": "pong",
	})
}
