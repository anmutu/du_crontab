/*
  author='du'
  date='2020/6/1 15:09'
*/
package worker

import (
	"context"
	"du_corntab/crontab/common"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type LogSink struct {
	client        *mongo.Client
	logCollection *mongo.Collection
	logChan       chan *common.JobLog
}

var (
	G_LogSink *LogSink
)

//初始化日志模块
func InitLogSink() (err error) {
	var (
		client *mongo.Client
	)
	if client, err = mongo.Connect(context.TODO(), options.Client().ApplyURI(G_config.MongodbUri)); err != nil {
		fmt.Println(err)
		return
	}
	G_LogSink = &LogSink{
		client:        client,
		logCollection: client.Database("cron").Collection("log"),
		logChan:       make(chan *common.JobLog, 1000),
	}

	//开启一个日志数据到mongdb的一个协程。
	go G_LogSink.writeLoop()

	return
}

func (logSink *LogSink) writeLoop() {
	var (
		log *common.JobLog
	)
	for {
		select {
		case log = <-logSink.logChan:
			//把日志写到mongdb中。
			log = log

		}
	}

}
