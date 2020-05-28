/*
  author='du'
  date='2020/5/28 7:17'
*/
package main

import (
	"context"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"time"
)

func main() {
	var (
		client         *clientv3.Client
		lease          clientv3.Lease
		leaseGrantResp *clientv3.LeaseGrantResponse
		leaseId        clientv3.LeaseID
		kv             clientv3.KV
		putResp        *clientv3.PutResponse
		getResp        *clientv3.GetResponse
	)

	//客户端配置
	config := clientv3.Config{
		Endpoints:   []string{"129.211.78.6:2379"},
		DialTimeout: 5 * time.Second,
	}

	//建立连接
	client, err := clientv3.New(config)
	if err != nil {
		fmt.Printf("连接失败：%s", err)
		return
	}

	//申请一个租约
	lease = clientv3.NewLease(client)

	//申请一个5s的租约。
	if leaseGrantResp, err = lease.Grant(context.TODO(), 5); err != nil {
		fmt.Println(err)
		return
	}
	leaseId = leaseGrantResp.ID

	//获取kv,然后Put kv,将之和租约关联起来，实现过期的效果
	kv = clientv3.KV(client)
	if putResp, err = kv.Put(context.TODO(), "/cron/lock/job1", "", clientv3.WithLease(leaseId)); err != nil {
		return
	}
	fmt.Println("写入成功,当前revision是:", putResp.Header.Revision)

	//模拟数据,每1s去get一下数据，看5s后数据有无过期
	for {
		if getResp, err = kv.Get(context.TODO(), "/cron/lock/job1"); err != nil {
			fmt.Println(err)
			return
		}
		if getResp.Count == 0 {
			fmt.Println("未获取到数据，已经过期了。")
			break
		}
		fmt.Println("还木有过期，当前数据", getResp.Kvs)
		time.Sleep(1 * time.Second)
	}

}
