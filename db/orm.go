package db

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/plugin/dbresolver"

	"code.byted.org/qiumingliang.123/personal_bot/conf"
)

var db *gorm.DB
var dbRead *gorm.DB

func InitMysqlConn() {
	dbRead = getDBClient("personal_bot")
	db = dbRead.Clauses(dbresolver.Write).Session(&gorm.Session{})
}

func getDBClient(dbName string) *gorm.DB {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%v)/%s?charset=utf8mb4&parseTime=True&loc=Local", conf.DbConf.User, conf.DbConf.Pass, conf.DbConf.Addr, conf.DbConf.Port, dbName)
	logrus.Infof("dsn: %v", dsn)
	DB, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info), // 打印所有sql
	})
	if err != nil {
		fmt.Printf("connect db(addr= %v) err: %v", dbName, err)
		panic("failed to connect database")
	}

	return DB
}
