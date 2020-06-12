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
	"time"
)

type LogSink struct {
	client         *mongo.Client
	logCollection  *mongo.Collection
	logChan        chan *common.JobLog
	autoCommitChan chan *common.LogBatch
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
	fmt.Println("enter InitLogSink")
	G_LogSink = &LogSink{
		client:         client,
		logCollection:  client.Database("cron").Collection("log"),
		logChan:        make(chan *common.JobLog, 1000),
		autoCommitChan: make(chan *common.LogBatch, 1000),
	}

	//开启一个日志数据到mongodb的一个协程。
	go G_LogSink.writeLoop()

	return
}

//把日志写入到mongodb的函数。
func (logSink *LogSink) writeLoop() {
	var (
		log          *common.JobLog
		logBatch     *common.LogBatch
		commitTimer  *time.Timer
		timeOutBatch *common.LogBatch //超时的批次
	)
	for {
		select {
		case log = <-logSink.logChan:
			if logBatch == nil {
				logBatch = &common.LogBatch{}
				//因为会有2个协程在操作这个batch,所以需要作一个串行化处理。
				//让这个回调函数发一个通知给writeLoop
				commitTimer = time.AfterFunc(time.Duration(G_config.JobLogCommitTimeout)*time.Millisecond,
					func(batch *common.LogBatch) func() {
						return func() {
							logSink.autoCommitChan <- batch
						}
					}(logBatch),
				)
			}
			logBatch.Logs = append(logBatch.Logs, log)
			if len(logBatch.Logs) >= G_config.JobLogBatchSize {
				//批次满了则要将这批次的日志存到mongodb,保存后要将batch清空。
				fmt.Println("5. writeLoop：日志批次已满。")
				logSink.saveLogs(logBatch)
				logBatch = nil
				commitTimer.Stop() //可能取消不了，那么下面的情况就要判断两个批次是否相等了。
			}
		case timeOutBatch = <-logSink.autoCommitChan:
			fmt.Println("5. writeLoop：日志批次未满。")
			if timeOutBatch != logBatch {
				continue
			}
			//把超时的批次写入到mongodb中，且保存后要将batch清空
			logSink.saveLogs(timeOutBatch)
			//timeOutBatch = nil
			logBatch = nil
		}
	}
}

//批量写入日志的函数。
func (logSink *LogSink) saveLogs(logs *common.LogBatch) {
	fmt.Println("5. saveLogs：将提交日志数据到mongodb里去了。 ")
	logSink.logCollection.InsertMany(context.TODO(), logs.Logs)
}

//scheduler模块调用，处理后将日志信息发送过来。
func (logSink *LogSink) Append2LogChan(jobLog *common.JobLog) {
	select {
	case logSink.logChan <- jobLog:
	default:
		//满了就丢弃了。
	}
}
