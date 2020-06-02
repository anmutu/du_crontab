/*
  author='du'
  date='2020/6/2 3:44'
*/
package main

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	var (
		client     *mongo.Client
		err        error
		database   *mongo.Database
		collection *mongo.Collection
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

	fmt.Printf("成功连接到了%s库里的%s", database.Name(), collection.Name())
}
