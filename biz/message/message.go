package message

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"code.byted.org/qiumingliang.123/personal_bot/biz/common"
	"code.byted.org/qiumingliang.123/personal_bot/biz/model"
)

var (
	createMessageURL     = "https://open.feishu.cn/open-apis/im/v1/messages"
	uploadImageURL       = "https://open.feishu.cn/open-apis/im/v1/images"
	getMessageHistoryURL = "https://open.feishu.cn/open-apis/im/v1/messages"
)

func SendAlertMessage(ctx context.Context, token, msgType string, chatID string) error {
	var err error
	if token == "" {
		token, err = common.GetTenantAccessToken(ctx)
		if err != nil {
			logrus.WithError(err).Errorf("failed to get tenant access token")
			return err
		}
	}

	var createResp *model.MessageItem
	var createReq *model.CreateMessageRequest
	switch msgType {
	case "text":
		content := "{\"text\":\"<at user_id=\\\"all\\\">所有人</at> 请注意，线上服务发生报警，请及时处理。 \\n服务负责人：<at user_id=\\\"ou_ba44c2d64d161c0f12d8548bef215311\\\">张三</at> \"}"
		createReq = GenCreateMessageRequest(ctx, chatID, content, msgType)

	case "post":
		content := "{\"zh_cn\":{\"title\":\"线上服务报警通知！\",\"content\":[[{\"tag\":\"at\",\"user_id\":\"all\",\"user_name\":\"所有人\"},{\"tag\":\"text\",\"text\":\"请注意，线上服务发生报警，请及时处理。\"}],[{\"tag\":\"text\",\"text\":\"服务负责人：\"},{\"tag\":\"at\",\"user_id\":\"ou_ba44c2d64d161c0f12d8548bef215311\",\"user_name\":\"张三\"}]]}}"
		createReq = GenCreateMessageRequest(ctx, chatID, content, msgType)
	case "interactive":
		image, err := UploadImage(ctx, token)
		if err != nil {
			logrus.WithError(err).Errorf("failed to upload image")
			return err
		}
		cardContent := ConstructAlterCard(ctx, image.ImageKey)
		createReq = GenCreateMessageRequest(ctx, chatID, cardContent, msgType)
	default:
		return nil
	}

	createResp, err = SendMessage(ctx, token, createReq)
	if err != nil {
		logrus.WithError(err).Errorf("send %v message failed, chat_id: %v", msgType, chatID)
		return err
	}

	msgID := createResp.MessageID
	logrus.Infof("succeed send alert message, msg_id: %v", msgID)
	return nil
}

func SendMessage(ctx context.Context, token string, createReq *model.CreateMessageRequest) (*model.MessageItem, error) {
	var err error
	if token == "" {
		token, err = common.GetTenantAccessToken(ctx)
		if err != nil {
			logrus.WithError(err).Errorf("failed to get tenant access token")
			return nil, err
		}
	}
	cli := &http.Client{}

	reqBytes, err := json.Marshal(createReq)
	if err != nil {
		logrus.WithError(err).Errorf("failed to marshal")
		return nil, err
	}
	req, err := http.NewRequest("POST", createMessageURL, strings.NewReader(string(reqBytes)))
	if err != nil {
		logrus.WithError(err).Errorf("new request failed")
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	q := req.URL.Query()
	q.Add("receive_id_type", "chat_id")
	req.URL.RawQuery = q.Encode()

	var logID string
	resp, err := cli.Do(req)
	if err != nil {
		return nil, fmt.Errorf("create message failed, err=%v", err)
	}
	if resp != nil && resp.Header != nil {
		logID = resp.Header.Get("x-tt-logid")
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logrus.WithError(err).Error("read body failed")
		return nil, err
	}

	createMessageResp := &model.CreateMessageResponse{}
	err = json.Unmarshal(body, createMessageResp)
	if err != nil {
		logrus.WithError(err).Errorf("failed to unmarshal")
		return nil, err
	}
	if createMessageResp.Code != 0 {
		logrus.Warnf("failed to create message, code: %v, msg: %v, log_id: %v", createMessageResp.Code, createMessageResp.Message, logID)
		return nil, fmt.Errorf("create message failed")
	}
	logrus.Infof("succeed create message, msg_id: %v", createMessageResp.Data.MessageID)
	return createMessageResp.Data, nil
}

func GenCreateMessageRequest(ctx context.Context, chatID, content, msgType string) *model.CreateMessageRequest {
	return &model.CreateMessageRequest{
		ReceiveID: chatID,
		Content:   content,
		MsgType:   msgType,
	}
}

func ConstructAlterCard(ctx context.Context, img string) (card string) {
	cardContent := &model.CardContent{
		Config: &model.CardConfig{
			WideScreenMode: true,
		},
		Header: &model.CardHeader{
			Template: "red",
			Title: &model.CardText{
				Tag:     "plain_text",
				Content: "1 级报警 - 数据平台",
			},
		},
	}
	elements := make([]interface{}, 0)
	// card block 1
	element1 := &model.CardElement{
		Tag: "div",
		Fields: []*model.CardField{
			{
				IsShort: true,
				Text: &model.CardText{
					Content: "**🕐 时间：**\n2021-02-23 20:17:51",
					Tag:     "lark_md",
				},
			}, {
				IsShort: true,
				Text: &model.CardText{
					Content: "**🔢 事件 ID：**\n336720",
					Tag:     "lark_md",
				},
			}, {
				IsShort: true,
				Text: &model.CardText{
					Content: "**📋 项目：**\nQA 7",
					Tag:     "lark_md",
				},
			}, {
				IsShort: true,
				Text: &model.CardText{
					Content: "**👤 一级值班：**\n<at id=ou_ba44c2d64d161c0f12d8548bef215311>张三</at>",
					Tag:     "lark_md",
				},
			}, {
				IsShort: true,
				Text: &model.CardText{
					Content: "**👤 二级值班：**\n<at id=all>所有人</at>",
					Tag:     "lark_md",
				},
			},
		},
	}
	elements = append(elements, element1)

	// card block 2, image block
	element2 := &model.CardElement{
		Tag:    "img",
		ImgKey: img,
		Alt: &model.CardText{
			Content: " ",
			Tag:     "plain_text",
		},
		Title: &model.CardText{
			Content: "支付方式 支付成功率低于 50%：",
			Tag:     "lark_md",
		},
	}
	elements = append(elements, element2)

	// card block 3, note block
	element3 := model.CardNote{
		Tag: "note",
	}
	element3Elements := make([]interface{}, 0)
	element3Elements = append(element3Elements, &model.CardText{
		Content: "🔴 支付失败数  🔵 支付成功数",
		Tag:     "plain_text",
	})
	element3.Elements = element3Elements
	elements = append(elements, element3)

	// card action block
	element4 := &model.CardActionBlock{
		Tag: "action",
	}
	actions := make([]interface{}, 0)
	button := &model.CardButton{
		Tag: "button",
		Text: &model.CardText{
			Tag:     "plain_text",
			Content: "跟进处理",
		},
		Type:  "primary",
		Value: map[string]string{"key1": "value1"},
	}
	actions = append(actions, button)
	selectMenu := &model.CardSelectMenu{
		Tag: "select_static",
		PlaceHolder: &model.CardText{
			Content: "暂时屏蔽",
			Tag:     "plain_text",
		},
		Options: []*model.CardOption{
			{
				Text: &model.CardText{
					Content: "屏蔽10分钟",
					Tag:     "plain_text",
				},
				Value: "1",
			}, {
				Text: &model.CardText{
					Content: "屏蔽30分钟",
					Tag:     "plain_text",
				},
				Value: "2",
			}, {
				Text: &model.CardText{
					Content: "屏蔽1小时",
					Tag:     "plain_text",
				},
				Value: "3",
			}, {
				Text: &model.CardText{
					Content: "屏蔽24小时",
					Tag:     "plain_text",
				},
				Value: "4",
			},
		},
		Value: map[string]string{"key": "value"},
	}
	actions = append(actions, selectMenu)
	element4.Actions = actions
	elements = append(elements, element4)

	// card split line
	element5 := &model.CardSplitLine{
		Tag: "hr",
	}
	elements = append(elements, element5)

	// card
	element6 := &model.CardElement{
		Tag: "div",
		Text: &model.CardText{
			Content: "🙋🏼 [我要反馈误报](https://open.feishu.cn/) | 📝 [录入报警处理过程](https://open.feishu.cn/)",
			Tag:     "lark_md",
		},
	}
	elements = append(elements, element6)

	cardContent.Elements = elements

	cardBytes, err := json.Marshal(cardContent)
	if err != nil {
		logrus.WithError(err).Errorf("failed to marshal")
		return ""
	}
	logrus.Infof("card_content: %v", string(cardBytes))
	return string(cardBytes)
}

func UploadImage(ctx context.Context, token string) (*model.UploadImageResponseBody, error) {
	cli := &http.Client{}

	pwd, _ := os.Getwd()
	parent := filepath.Dir(pwd)
	path := parent + "/resource/upload/alert.png"
	logrus.Infof("path: %v", path)
	image, err := os.Open(path)
	if err != nil {
		logrus.WithError(err).Errorf("failed to open image")
		return nil, err
	}
	defer image.Close()

	buffer := &bytes.Buffer{}
	write := multipart.NewWriter(buffer)
	w, err := write.CreateFormFile("image", filepath.Base(path))
	if err != nil {
		logrus.WithError(err).Errorf("failed to create form file")
		return nil, err
	}
	_, err = io.Copy(w, image)
	if err != nil {
		logrus.WithError(err).Errorf("copy image failed")
		return nil, err
	}
	params := make(map[string]string)
	params["image_type"] = "message"
	for k, v := range params {
		err = write.WriteField(k, v)
		if err != nil {
			return nil, err
		}
	}

	err = write.Close()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", uploadImageURL, buffer)
	if err != nil {
		logrus.WithError(err).Errorf("new request failed")
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Content-Type", write.FormDataContentType())

	var logID string
	resp, err := cli.Do(req)
	if err != nil {
		return nil, fmt.Errorf("create message failed, err=%v", err)
	}
	if resp != nil && resp.Header != nil {
		logID = resp.Header.Get("x-tt-logid")
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logrus.WithError(err).Error("read body failed")
		return nil, err
	}
	uploadImageResp := &model.UploadImageResponse{}
	err = json.Unmarshal(body, uploadImageResp)
	if err != nil {
		logrus.WithError(err).Errorf("failed to unmarshal")
		return nil, err
	}
	if uploadImageResp.Code != 0 {
		logrus.Warnf("failed to upload image, code: %v, msg: %v, log_id: %v", uploadImageResp.Code, uploadImageResp.Message, logID)
		return nil, fmt.Errorf("create image failed")
	}
	return uploadImageResp.Data, nil
}

func GetChatAllMessageAndReview(ctx context.Context, token, chatID string) error {
	var err error
	if token == "" {
		token, err = common.GetTenantAccessToken(ctx)
		if err != nil {
			logrus.WithError(err).Errorf("failed to get tenant access token")
			return err
		}
	}
	start := "0"
	end := fmt.Sprintf("%v", time.Now().Unix())

	pwd, _ := os.Getwd()
	parent := filepath.Dir(pwd)
	path := parent + fmt.Sprintf("/resource/download/chat_%v_history.txt", chatID)

	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		logrus.WithError(err).Errorf("open file failed")
		return err
	}
	defer file.Close()

	write := bufio.NewWriter(file)
	write.WriteString(fmt.Sprintf("chat(%v) history\n", chatID))

	hasMore := true
	pageToken := ""
	for {
		if !hasMore {
			break
		}
		getMsgResp, err := GetChatMessageHistory(ctx, token, chatID, start, end, pageToken, "10")
		if err != nil {
			logrus.WithError(err).Errorf("failed to get chat message")
			break
		}

		if len(getMsgResp.Items) > 0 {
			for _, item := range getMsgResp.Items {
				senderID := item.Sender.ID
				createTime := item.CreateTime
				intCreateTime, err := strconv.ParseInt(createTime, 10, 64)
				if err != nil {
					continue
				}

				createTime = fmt.Sprintf("%v", time.Unix(intCreateTime/1000, 0))
				content := item.Body.Content
				str := fmt.Sprintf("chatter(%v) at (%v) send: %v", senderID, createTime, content)
				write.WriteString(str + "\n")
			}
			write.Flush()
		}
		pageToken = getMsgResp.PageToken
		hasMore = getMsgResp.HasMore
		time.Sleep(50 * time.Millisecond)
	}

	return nil
}

func GetChatMessageHistory(ctx context.Context, token, chatID string, start, end, pageToken, pageSize string) (*model.GetMessageHistoryBody, error) {
	var err error
	if token == "" {
		token, err = common.GetTenantAccessToken(ctx)
		if err != nil {
			logrus.WithError(err).Errorf("failed to get tenant access token")
			return nil, err
		}
	}
	cli := &http.Client{}

	req, err := http.NewRequest("GET", getMessageHistoryURL, nil)
	if err != nil {
		logrus.WithError(err).Errorf("new request failed")
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	q := req.URL.Query()
	q.Add("container_id_type", "chat")
	q.Add("container_id", chatID)
	q.Add("start_time", start)
	q.Add("end_time", end)
	q.Add("page_token", pageToken)
	q.Add("page_size", pageSize)
	req.URL.RawQuery = q.Encode()

	var logID string
	resp, err := cli.Do(req)
	if err != nil {
		return nil, fmt.Errorf("get message failed, err=%v", err)
	}
	if resp != nil && resp.Header != nil {
		logID = resp.Header.Get("x-tt-logid")
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logrus.WithError(err).Error("read body failed")
		return nil, err
	}

	getMessageResp := &model.GetMessageHistoryResponse{}
	err = json.Unmarshal(body, getMessageResp)
	if err != nil {
		logrus.WithError(err).Errorf("failed to unmarshal")
		return nil, err
	}
	if getMessageResp.Code != 0 {
		logrus.Warnf("failed to get message, code: %v, msg: %v, log_id: %v", getMessageResp.Code, getMessageResp.Message, logID)
		return nil, fmt.Errorf("get message hitory failed")
	}

	return getMessageResp.Data, nil
}
