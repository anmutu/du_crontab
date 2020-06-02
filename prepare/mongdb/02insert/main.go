/*
  author='du'
  date='2020/6/2 9:01'
*/
package main

import (
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/net/context"
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
		JobName:   "test1",
		Command:   "echo hi",
		Err:       "",
		Content:   "hi",
		StartTime: time.Now().Unix(),
		EndTime:   time.Now().Unix() + 10,
	}

	if _, err = collection.InsertOne(context.TODO(), logRecord); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("成功将数据插入到db中")

}
