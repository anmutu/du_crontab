/*
  author='du'
  date='2020/6/2 23:49'
*/
package master

import (
	"context"
	"du_corntab/crontab/common"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"
)

type LogMgr struct {
	client        *mongo.Client
	logCollection *mongo.Collection
}

var (
	G_LogMgr *LogMgr
)

func InitLogMgr() (err error) {
	var (
		client *mongo.Client
	)

	if client, err = mongo.Connect(context.TODO(), options.Client().ApplyURI(G_config.MongodbUri)); err != nil {
		fmt.Println(err)
		return
	}
	G_LogMgr = &LogMgr{
		client:        client,
		logCollection: client.Database("cron").Collection("log"),
	}
	return
}

// 查看任务日志
func (logMgr *LogMgr) GetLoglist(name string, skip int, limit int) (logArr []*common.JobLog, err error) {
	var (
	//filter *common.JobLogFilter
	//logSort *common.SortLogByStartTime
	//cursor *mongo.Cursor
	//jobLog *common.JobLog
	)

	// len(logArr)
	logArr = make([]*common.JobLog, 0)

	// 过滤条件
	//filter = &common.JobLogFilter{JobName: name}
	//filter = &common.JobLogFilter{}

	// 按照任务开始时间倒排
	//logSort = &common.SortLogByStartTime{SortOrder: -1}

	// 查询
	//ops := options.Find().SetSort(logSort).SetLimit(int64(limit)).SetSkip(int64(skip))
	//ops = ops

	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)

	opts := options.Find().SetSort(bson.D{{"scheduleTime", -1}}).SetLimit(100)
	cur, err := logMgr.logCollection.Find(context.TODO(), bson.D{{"jobName", name}}, opts)

	if err != nil {
		log.Fatal(err)
	}
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		//var result bson.M
		jobLog := &common.JobLog{}
		err := cur.Decode(jobLog)
		if err != nil {
			log.Fatal(err)
		}
		logArr = append(logArr, jobLog)
	}
	if err := cur.Err(); err != nil {
		log.Fatal(err)
	}
	return
}
