package chat

import (
	"context"
	"testing"

	"github.com/sirupsen/logrus"

	"code.byted.org/qiumingliang.123/personal_bot/biz/common"
	"code.byted.org/qiumingliang.123/personal_bot/biz/model"
)

func TestCreateChatAndInviteUser(t *testing.T) {
	ctx := context.Background()

	_, err := CreateChatAndInviteUser(ctx, "")
	if err != nil {
		t.Error("test failed.")
	} else {
		t.Log("create chat and invite user succeed.")
	}

}

func TestUpdateChat(t *testing.T) {
	ctx := context.Background()

	token, err := common.GetTenantAccessToken(ctx)
	if err != nil {
		logrus.WithError(err).Errorf("failed to get tenant access token")
		t.Fail()
		return
	}
	chatID, err := CreateChatAndInviteUser(ctx, "")
	if err != nil {
		t.Error("test failed.")
	} else {
		t.Log("create chat and invite user succeed.")
	}
	chatInfo, err := GetChatInfo(ctx, token, chatID)
	if err != nil {
		logrus.WithError(err).Errorf("failed to get chat")
		t.Fail()
	}
	logrus.Infof("chat_info: %v", chatInfo)

	updateChatRequest := &model.UpdateChatRequest{
		Name: "（已解决）" + chatInfo.Name,
	}
	if chatInfo.I18nNames != nil {
		i18nName := &model.I18nNames{
			ZhCn: "（已解决）" + chatInfo.I18nNames.ZhCn,
			EnUs: "(solved)" + chatInfo.I18nNames.EnUs,
			JaJp: "（解決済み）" + chatInfo.I18nNames.JaJp,
		}
		updateChatRequest.I18nNames = i18nName
	}

	_, err = UpdateChat(ctx, token, chatID, updateChatRequest)
	if err != nil {
		logrus.WithError(err).Errorf("failed to update chat")
		t.Fail()
	}
}
