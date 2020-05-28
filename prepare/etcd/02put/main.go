/*
  author='du'
  date='2020/5/25 11:33'
*/
package main

import (
	"context"
	"fmt"
	"go.etcd.io/etcd/clientv3"
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
	putResp, err := kv.Put(context.TODO(), "/cron/jobs/job1", "hello world1", clientv3.WithPrevKV())
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("之前的版本号是:", putResp.Header.Revision)
		if putResp.PrevKv != nil {
			fmt.Println("PrevKv的值是：", string(putResp.PrevKv.Value))
		} else {
			fmt.Println("第一次put值哦")
		}
	}

}
