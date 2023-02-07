package db

import (
	"context"
	"fmt"
	"reflect"

	"github.com/sirupsen/logrus"

	"code.byted.org/qiumingliang.123/personal_bot/biz/model"
)

func GetFindData(a interface{}) *model.CronData {
	if a == nil {
		return nil
	}
	logrus.Infof("data: %v, type: %v", a, reflect.TypeOf(a).String())
	if reflect.TypeOf(a).String() == "*model.CronData" {
		data := a.(*model.CronData)
		return &model.CronData{
			ChatID:    data.ChatID,
			StartTime: data.StartTime,
			Intervals: data.Intervals,
			Rank:      data.Rank,
			Content:   data.Content,
			MsgType:   data.MsgType,
		}
	} else if reflect.TypeOf(a).String() == "*model.JobMap" {
		data := a.(*model.JobMap)
		return &model.CronData{
			ChatID:    data.ChatID,
			StartTime: data.Start,
			Intervals: data.Gap,
			Rank:      data.Rank,
			Content:   data.Content,
			MsgType:   data.Type,
		}
	}
	return nil
}

func PutCronData(ctx context.Context, data *model.CronData) error {
	log := logrus.WithContext(ctx)
	tx := db.WithContext(ctx).Table("cron_job")
	find_data := GetFindData(data)
	if find_data == nil {
		log.Errorf("find err")
		return fmt.Errorf("put error")
	}
	ret := make([]*model.CronData, 0)
	err := tx.Where(find_data).Find(&ret).Error
	if err == nil && len(ret) == 1 {
		return OpenCronJob(ctx, data)
	}

	err = tx.Where(find_data).FirstOrCreate(data).Error
	if err != nil {
		log.WithError(err).Errorf("Failed to insert cron data, data: %v", data)
		return err
	}
	return nil
}

func GetCronData(ctx context.Context) ([]*model.CronData, error) {
	log := logrus.WithContext(ctx)
	tx := dbRead.WithContext(ctx).Table("cron_job")
	ret := make([]*model.CronData, 0)
	err := tx.Where(&model.CronData{IsDelete: false}).Find(&ret).Error
	if err != nil {
		log.WithError(err).Errorf("Failed to get cron data")
		return nil, err
	}
	return ret, nil
}

func CloseCronJob(ctx context.Context, data *model.CronData) error {
	log := logrus.WithContext(ctx)
	tx := db.WithContext(ctx).Table("cron_job")
	find_data := GetFindData(data)
	if find_data == nil {
		log.Errorf("find err")
		return fmt.Errorf("put error")
	}
	err := tx.Where(find_data).Update("is_open", false).Error
	if err != nil {
		log.WithError(err).Errorf("Failed to insert cron data, data: %v", data)
		return err
	}
	return nil
}

func OpenCronJob(ctx context.Context, data *model.CronData) error {
	log := logrus.WithContext(ctx)
	tx := db.WithContext(ctx).Table("cron_job")
	find_data := GetFindData(data)
	if find_data == nil {
		log.Errorf("find err")
		return fmt.Errorf("put error")
	}
	err := tx.Where(find_data).Update("is_open", true).Error
	if err != nil {
		log.WithError(err).Errorf("Failed to insert cron data, data: %v", data)
		return err
	}
	return nil
}
