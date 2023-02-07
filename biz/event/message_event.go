package event

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"

	"code.byted.org/qiumingliang.123/personal_bot/biz/message/solve"
	"code.byted.org/qiumingliang.123/personal_bot/biz/model"
)

func removeEvent(ctx context.Context, s string) string {
	str := ""
	for i := 1; i < len(s)-1; i++ {
		str += fmt.Sprintf("%c", s[i])
	}
	return str
}

func HandleReceiveMessageEvent(ctx context.Context, event *model.ReceiveMessageEvent) error {
	msg := event.Event.Message

	if len(event.Event.Message.Mentions) == 0 {
		logrus.Infof("At All message, ignore")
		return nil
	}

	switch msg.MessageType {
	case "text":
		return solve.TextMessage(ctx, event)
	case "post":

	default:
		logrus.Infof("unhandled message type, msg_type: %v", msg.MessageType)
	}
	return nil
}
