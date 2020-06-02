/*
  author='du'
  date='2020/6/2 10:09'
*/
package main

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type LogRecord struct {
	JobName   string `bson:"jobName"` // 任务名
	Command   string `bson:"command"` // shell命令
	Err       string `bson:"err"`     // 脚本错误
	Content   string `bson:"content"`
	StartTime int64  `bson:"starttime"`
	EndTime   int64  `bson:"endtime"`
}

func main() {
	var (
		client     *mongo.Client
		err        error
		database   *mongo.Database
		collection *mongo.Collection
		logRecord  *LogRecord
		logRecord1 *LogRecord
		logArr     []interface{}
	)
	// 1, 建立连接
	if client, err = mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://129.211.78.6:27017")); err != nil {
		fmt.Println(err)
		return
	}

	// 2, 选择数据库my_db
	database = client.Database("my_db")

	// 3, 选择表my_collection
	collection = database.Collection("my_collection")
	fmt.Sprintf("成功连接到了%s库里的%s", database.Name(), collection.Name())

	logRecord = &LogRecord{
		JobName:   "test1_batch",
		Command:   "echo hi",
		Err:       "",
		Content:   "hi",
		StartTime: time.Now().Unix(),
		EndTime:   time.Now().Unix() + 10,
	}

	logRecord1 = &LogRecord{
		JobName:   "test2_batch",
		Command:   "echo hi",
		Err:       "",
		Content:   "hi",
		StartTime: time.Now().Unix(),
		EndTime:   time.Now().Unix() + 10,
	}

	logArr = []interface{}{logRecord, logRecord1}
	if _, err = collection.InsertMany(context.TODO(), logArr); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("成功将批量数据插入到db中")

}
