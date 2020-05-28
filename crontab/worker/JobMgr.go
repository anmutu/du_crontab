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
	//"go.etcd.io/etcd/mvcc/mvccpb"
	"github.com/coreos/etcd/mvcc/mvccpb"
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
func InitJobMgr() (err error) {
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

	//启动监听
	G_JobMgr.watchJobs()

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
		watchResp            clientv3.WatchResponse
		watchEvent           *clientv3.Event
		job                  *common.Job
		jobName              string
		jobEvent             *common.JobEvent
	)
	if getResp, err = JobMgr.kv.Get(context.TODO(), common.JOB_SAVE_DIR, clientv3.WithPrevKV()); err != nil {
		return
	} else {
		//得到当前的所有任务
		for _, kvPairs := range getResp.Kvs {
			if job, err := common.UnpackJob(kvPairs.Value); err != nil {
				jobEvent = common.BuildJobEvent(common.JOB_EVENT_SAVE, job)
				//TODO：把这个job同步给scheduler这个调度协程
				fmt.Println(*jobEvent)
			}
		}
	}

	//从此revision向后监听变化事件
	go func() {
		watcherStartRevision = getResp.Header.Revision + 1
		watchChan = JobMgr.watcher.Watch(context.TODO(), common.JOB_SAVE_DIR, clientv3.WithRev(watcherStartRevision), clientv3.WithPrefix())
		//拿到watchChan就可以监听了
		for watchResp = range watchChan {
			for _, watchEvent = range watchResp.Events {
				switch watchEvent.Type {
				case mvccpb.PUT: //说明这里是任务保存事件
					if job, err = common.UnpackJob(watchEvent.Kv.Value); err != nil {
						continue //如果不能正常转换就忽略
					}
					//构造一个更新的event
					jobEvent = common.BuildJobEvent(common.JOB_EVENT_SAVE, job)
				case mvccpb.DELETE:
					jobName = common.ExtractJobName(string(watchEvent.Kv.Key))
					//构造一个删除的event
					job = &common.Job{Name: jobName}
					jobEvent = common.BuildJobEvent(common.JOB_EVENT_DELETE, job)

				}
				//TODO 把jobEvent推给scheduler.
				fmt.Println(*jobEvent)
			}
		}
	}()
	return
}
