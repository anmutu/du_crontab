/*
  author='du'
  date='2020/5/27 23:46'
*/
package worker

import (
	"context"
	"du_corntab/crontab/common"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"time"
)

//任务管理器
type JobMgr struct {
	client  *clientv3.Client //这里用的是指针
	kv      clientv3.KV
	lease   clientv3.Lease
	watcher clientv3.Watcher
}

//worker是监听到etcd里的任务，然后把任务同步到内存里。

var (
	G_JobMgr *JobMgr //单例对象,指针
)

//初始化管理器
func InitJogMgr() (err error) {
	var (
		config  clientv3.Config
		client  *clientv3.Client
		kv      clientv3.KV
		lease   clientv3.Lease
		watcher clientv3.Watcher
	)

	//初始化配置
	config = clientv3.Config{
		Endpoints:   G_config.EtcdEndpoints,                                     //集群地址
		DialTimeout: time.Duration(G_config.EtcdDialTimeout) * time.Millisecond, //连接超时时间
	}

	//建立连接
	if client, err = clientv3.New(config); err != nil {
		return
	}

	//得到kv,lease租约和watcher监听器
	kv = clientv3.NewKV(client)
	lease = clientv3.NewLease(client)
	watcher = clientv3.Watcher(client)

	//将之赋值给单例对像
	G_JobMgr = &JobMgr{
		client:  client,
		kv:      kv,
		lease:   lease,
		watcher: watcher,
	}
	fmt.Println("初始化任务管理器成功。")
	return
}

//监听任务的变化
//就是通过watch的api去监听etcd
//思路就是get得到某目录下的所有任务，并拿到当前集群的revision
//接着就从此revision向后监听变化
func (JobMgr *JobMgr) watchJobs() (err error) {
	var (
		getResp              *clientv3.GetResponse
		watcherStartRevision int64
		watchChan            clientv3.WatchChan
		//watchResp clientv3.WatchResponse
	)
	if getResp, err = JobMgr.kv.Get(context.TODO(), common.JOB_SAVE_DIR, clientv3.WithPrevKV()); err != nil {
		return
	} else {
		//得到当前的所有任务
		for _, kvPairs := range getResp.Kvs {
			if job, err := common.UnpackJob(kvPairs.Value); err != nil {
				//TODO：把这个job同步给scheduler这个调度协程
				job = job
			}
		}
	}

	//从此revision向后监听变化事件
	go func() {
		watcherStartRevision = getResp.Header.Revision + 1
		watchChan = JobMgr.watcher.Watch(context.TODO(), common.JOB_SAVE_DIR, clientv3.WithRev(watcherStartRevision))

	}()

}
