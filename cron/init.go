package cron

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jasonlvhit/gocron"
	"github.com/sirupsen/logrus"

	"code.byted.org/qiumingliang.123/personal_bot/biz/message"
	"code.byted.org/qiumingliang.123/personal_bot/biz/model"
	"code.byted.org/qiumingliang.123/personal_bot/db"
)

var job_map map[string]map[string]bool
var is_open_map map[string]bool
var job_list map[string][]string
var func_map map[string]func()
var mutex sync.Mutex
var func_exist map[string]bool

type cronMap struct {
	ChatNums map[string]int64
}

func InitCronLoop() {
	gocron.Start()
	job_map = make(map[string]map[string]bool)
	is_open_map = make(map[string]bool)
	job_list = make(map[string][]string)
	func_map = make(map[string]func())
	func_exist = make(map[string]bool)
	go init_jobs_from_db()
}

func init_jobs_from_db() {
	ctx := context.Background()
	datas, err := db.GetCronData(ctx)
	if err != nil {
		return
	}
	for _, data := range datas {
		PutLoopJob(ctx, data)
	}
}

func getkeyfromjob(job *model.JobMap) string {
	return fmt.Sprintf("%v^-:-^%v^-:-^%v^-:-^%v^-:-^%v^-:-^%v", job.Gap, job.ChatID, job.Content, job.Type, job.Start, job.Rank)
}

func Getjobfromkey(str string) *model.JobMap {
	ss := strings.Split(str, "^-:-^")
	if len(ss) != 6 {
		return nil
	}
	job := &model.JobMap{}
	job.Gap, _ = strconv.Atoi(ss[0])
	job.ChatID = ss[1]
	job.Content = ss[2]
	job.Type = ss[3]
	job.Start = ss[4]
	job.Rank = ss[5]
	return job
}

func AsyncSendJob(job *model.JobMap) {
	var err error
	req := message.GenCreateMessageRequest(context.Background(), job.ChatID, "{\"text\":\""+job.Content+"\"}", "text")
	logrus.Infof("send content: %v", req.Content)
	for i := 0; i < 3; i++ {
		if _, err = message.SendMessage(context.Background(), "", req); err == nil {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func CloseJob(chatID string, num int) {
	mutex.Lock()
	defer mutex.Unlock()
	is_open_map[job_list[chatID][num-1]] = false
	gocron.Remove(func_map[job_list[chatID][num-1]])
	func_exist[job_list[chatID][num-1]] = false
	db.CloseCronJob(context.Background(), db.GetFindData(Getjobfromkey(job_list[chatID][num-1])))
}

func OpenJob(chatID string, num int) {
	mutex.Lock()
	defer mutex.Unlock()
	is_open_map[job_list[chatID][num-1]] = true
	job := Getjobfromkey(job_list[chatID][num-1])
	gocron.Every(1).Day().At(job.Start).Do(do_func_cron, job)
}

func GetJobMap(chatID string) map[string]bool {
	return job_map[chatID]
}

func GetJobList(chatID string) []string {
	return job_list[chatID]
}

func GetIsOpen() map[string]bool {
	return is_open_map
}

func do_func_cron(f func(), job *model.JobMap) {
	key := getkeyfromjob(job)
	v, ok := func_exist[key]
	if ok && v {
		return
	}
	func_exist[key] = true
	if job.Rank == "m" {
		gocron.Every(uint64(job.Gap)).Minutes().Do(func_map[getkeyfromjob(job)])
	} else if job.Rank == "h" {
		gocron.Every(uint64(job.Gap)).Hours().Do(func_map[getkeyfromjob(job)])
	}
}

func PutLoopJob(ctx context.Context, data *model.CronData) error {
	if data == nil {
		return nil
	}

	job := &model.JobMap{
		Gap:     data.Intervals,
		ChatID:  data.ChatID,
		Content: data.Content,
		Type:    data.MsgType,
		Start:   data.StartTime,
		Rank:    data.Rank,
		IsOpen:  data.IsOpen,
	}

	mutex.Lock()
	defer mutex.Unlock()
	putJobMap(job)

	if data.Rank == "h" {
		logrus.Infof("add Hour job, time: %v, gap: %v%v", data.StartTime, job.Gap, data.Rank)
		func_map[getkeyfromjob(job)] = func() {
			AsyncSendJob(job)
		}

		if data.IsOpen {
			gocron.Every(1).Day().At(data.StartTime).Do(do_func_cron, job)
		}
	}

	if data.Rank == "m" {
		logrus.Infof("add Minute job, time: %v, gap: %v%v", data.StartTime, job.Gap, data.Rank)
		func_map[getkeyfromjob(job)] = func() {
			AsyncSendJob(job)
		}

		if data.IsOpen {
			gocron.Every(1).Day().At(data.StartTime).Do(do_func_cron, job)
		}
	}
	return nil
}

func putJobMap(job *model.JobMap) {
	if _, ok := job_map[job.ChatID]; !ok {
		job_map[job.ChatID] = make(map[string]bool)
		job_list[job.ChatID] = make([]string, 0)
	}
	job_map[job.ChatID][getkeyfromjob(job)] = true
	job_list[job.ChatID] = append(job_list[job.ChatID], getkeyfromjob(job))
	if job.IsOpen {
		is_open_map[getkeyfromjob(job)] = true
	}
	logrus.Infof("key: %s", getkeyfromjob(job))
}
