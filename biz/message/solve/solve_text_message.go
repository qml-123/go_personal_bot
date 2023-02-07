package solve

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"github.com/sirupsen/logrus"

	"code.byted.org/qiumingliang.123/personal_bot/biz/common"
	"code.byted.org/qiumingliang.123/personal_bot/biz/model"
	"code.byted.org/qiumingliang.123/personal_bot/cron"
	"code.byted.org/qiumingliang.123/personal_bot/db"
)

func getVal(s []rune, i int) ([]rune, []rune, int) {
	j := i
	k, v, fir := make([]rune, 0), make([]rune, 0), true
	for ; j < len(s); j++ {
		if s[j] == ' ' || s[j] == '\n' {
			break
		}
		if fir == true && s[j] == ':' {
			fir = false
			continue
		}
		if fir {
			k = append(k, s[j])
		} else {
			v = append(v, s[j])
		}
	}
	return k, v, j - i
}

func removeAt(ss string) string {
	ss = common.GetStrText(ss)
	s := []rune(ss)
	str := make([]rune, 0)
	i := 0

	if s[0] != '@' {
		return string(s)
	}
	f := false
	for ; i < len(s); i++ {
		if !f && (s[i] == ' ' || s[i] == '\n') {
			f = true
		}
		if f {
			str = append(str, s[i])
		}
	}
	return string(str)
}

func removeText(ss string) string {
	s := []rune(ss)
	str := make([]rune, 0)
	for i := 0; i < len(s); i++ {
		if s[i] == '"' {
			continue
		}
		str = append(str, s[i])
	}
	st := common.GetStrText(string(str))
	return st[5:]
}

type Text struct {
	text string `json:"text"`
}

func TextMessage(ctx context.Context, event *model.ReceiveMessageEvent) error {
	log := logrus.WithContext(ctx)
	//t := Text{}
	var tempMap map[string]interface{}
	if err := json.Unmarshal([]byte(event.Event.Message.Content), &tempMap); err != nil {
		logrus.Errorf("json Unmarshal err, content: %v", event.Event.Message.Content)
		return err
	}
	log.Infof("get_message:%v", tempMap)

	content := []rune(common.GetStrText(removeAt(tempMap["text"].(string))))
	contentStr := string(content)

	if common.IsEqualString(contentStr, "叫") {
		return SendJobList(ctx, event)
	}

	if len(content) <= 1 {
		log.Warnf("len <= 1")
		return nil
	}

	if solveClose(ctx, content, event.Event.Message.ChatID) {
		return nil
	}

	if content[0] != '叫' {
		return nil
	}

	m := make(map[string]interface{})
	for i := 1; i < len(content); {
		if content[i] == ' ' {
			i++
			continue
		}
		k, v, offset := getVal(content, i)
		if offset == 1 {
			i++
			continue
		}
		if string(k) == "gap" {
			intNum, err := strconv.Atoi(string(v))
			if err != nil {
				return err
			}
			m[string(k)] = intNum
		} else {
			m[string(k)] = string(v)
		}
		i += offset
	}
	arr, err := json.Marshal(m)
	if err != nil {
		log.Errorf("json Marshal error, m: %v", m)
		return err
	}

	data := &model.CronData{}
	err = json.Unmarshal(arr, data)
	if err != nil {
		log.Errorf("json Unmarshal error, arr: %v", arr)
		return err
	}
	data.MsgType = event.Event.Message.MessageType
	data.OpenID = event.Event.Sender.SenderID["open_id"]
	data.IsOpen = true
	data.ChatID = event.Event.Message.ChatID
	log.Infof("get map: %v, data: %+v", m, data)
	if len(data.StartTime) != 5 || !unicode.IsNumber(rune(data.StartTime[0])) || !unicode.IsNumber(rune(data.StartTime[1])) || !unicode.IsNumber(rune(data.StartTime[3])) || !unicode.IsNumber(rune(data.StartTime[4])) {
		log.Errorf("start time error. value: %v", data.StartTime)
		return fmt.Errorf("time error")
	}
	if err = db.PutCronData(ctx, data); err != nil {
		return err
	}

	return cron.PutLoopJob(ctx, data)
}

func SendJobList(ctx context.Context, event *model.ReceiveMessageEvent) error {
	logrus.WithContext(ctx).Infof("SendJobList")
	m := cron.GetJobList(event.Event.Message.ChatID)
	is_open := cron.GetIsOpen()
	content := ""
	num := 0
	for _, v := range m {
		job := cron.Getjobfromkey(v)
		if job == nil {
			continue
		}
		num++
		content += fmt.Sprintf("%d", num) + ". start:" + job.Start + " gap:" + fmt.Sprintf("%d", job.Gap) + " rank:" + job.Rank + " content:" + job.Content + " is_open:" + strconv.FormatBool(is_open[v]) + "\\n"
	}
	if content == "" {
		content = "不叫"
	}
	cron.AsyncSendJob(&model.JobMap{
		ChatID:  event.Event.Message.ChatID,
		Content: content,
	})
	return nil
}

func solveClose(ctx context.Context, content []rune, chatID string) bool {
	logrus.WithContext(ctx).Infof("solve_close")
	k_list := strings.Split(string(content), " ")
	if len(k_list) != 2 || (k_list[0] != "close" && k_list[0] != "open") {
		return false
	}
	m := cron.GetJobList(chatID)
	v_list := strings.Split(k_list[1], ",")
	nums := make([]int, 0)
	for _, v := range v_list {
		v_num, err := strconv.Atoi(v)
		if err != nil {
			logrus.WithError(err).Errorf("atoi error")
			cron.AsyncSendJob(&model.JobMap{
				ChatID:  chatID,
				Content: "不叫",
			})
			return true
		}
		if v_num > len(m) || v_num <= 0 {
			logrus.WithError(err).Errorf("num over size")
			cron.AsyncSendJob(&model.JobMap{
				ChatID:  chatID,
				Content: "看不懂" + v,
			})
			return true
		}
		nums = append(nums, v_num)
	}
	for _, num := range nums {
		if k_list[0] == "open" {
			cron.OpenJob(chatID, num)
		} else if k_list[0] == "close" {
			cron.CloseJob(chatID, num)
		}
	}

	return true
}
