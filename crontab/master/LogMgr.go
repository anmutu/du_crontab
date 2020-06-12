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
	logArr = make([]*common.JobLog, 0)

	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)

	opts := options.Find().SetSort(bson.D{{"scheduleTime", -1}}).SetLimit(int64(limit)).SetSkip(int64(skip))
	cur, err := logMgr.logCollection.Find(context.TODO(), bson.D{{"jobName", name}}, opts)

	fmt.Println(name)
	fmt.Println(limit)
	fmt.Println(skip)
	fmt.Println(opts.Sort)

	if err != nil {
		log.Fatal(err)
	}
	defer cur.Close(ctx)

	for cur.Next(ctx) {
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
