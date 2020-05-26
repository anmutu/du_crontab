/*
  author='du'
  date='2020/5/26 0:35'
*/
package main

import (
	"context"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"time"
)

func main() {
	//客户端配置
	config := clientv3.Config{
		Endpoints:   []string{"129.211.78.6:2379"},
		DialTimeout: 25 * time.Second,
	}
	//建立连接
	client, err := clientv3.New(config)
	if err != nil {
		fmt.Printf("连接失败：%s", err)
		return
	}
	//键值对
	kv := clientv3.NewKV(client)
	if getResp, err := kv.Get(context.TODO(), "/cron/jobs/job1"); err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(getResp.Kvs, getResp.Count)
	}

}
