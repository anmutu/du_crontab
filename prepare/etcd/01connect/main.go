/*
  author='du'
  date='2020/5/25 10:09'
*/
package main

import (
	"fmt"
	"go.etcd.io/etcd/clientv3"
	"time"
)

func main() {
	//客户端配置
	config := clientv3.Config{
		Endpoints:   []string{"129.211.78.6:2379"},
		DialTimeout: 5 * time.Second,
	}
	//建立连接
	_, err := clientv3.New(config)
	if err != nil {
		fmt.Printf("连接失败：%s", err)
		return
	}
	print("连接成功")
}
