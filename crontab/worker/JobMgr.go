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

	//启动killer的监听
	G_JobMgr.watchKiller()

	fmt.Println("初始化任务管理器成功。监听了任务启动和killer的监听。")
	return
}

//监听任务的变化
//就是通过watch的api去监听etcd
//1.思路就是get得到某目录下的所有任务,将之给到Scheduler的channel
//2.拿到当前集群的revision，接着就从此revision向后监听变化,将之给到Scheduler的channel
//watchJobs的作用就是把目录里有的任务和目录后面版本的变化的job都推送到Scheduler的jobEventChan的channel里。
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
		kvPair               *mvccpb.KeyValue
	)
	if getResp, err = JobMgr.kv.Get(context.TODO(), common.JOB_SAVE_DIR, clientv3.WithPrefix()); err != nil {
		return
	}

	//得到当前的所有任务
	go func() {
		for _, kvPair = range getResp.Kvs {
			fmt.Println(getResp.Kvs)
			if job, err = common.UnpackJob(kvPair.Value); err == nil {
				jobEvent = common.BuildJobEvent(common.JOB_EVENT_SAVE, job)
				//把这个job同步给scheduler这个调度协程
				fmt.Println("watchJobs：从现有的jobs里将其任务发送给scheduler,job名为：", jobEvent.Job.Name)
				G_Scheduler.PushJobEvent(jobEvent)
			}
		}
	}()

	//从此revision向后监听变化事件，把变化的job相关信息发送到Scheduler的相关channel。
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
				//把jobEvent推给scheduler.就是把jobEvent给到它的channel。
				//fmt.Println("将要推送给sheduler的jobEvent的key是", jobEvent.Job.Name)
				fmt.Println("watchJobs：监控到有事件变化，将watch到的job发送给scheduler,job名为：", jobEvent.Job.Name)

				G_Scheduler.PushJobEvent(jobEvent)
			}
		}
	}()
	return
}

//监听强杀的目录
//还是跟监听job的目录一样，目的都是把相关信息推送到Scheduler。
func (jobMgr *JobMgr) watchKiller() {
	go func() {
		var (
			watchChan  clientv3.WatchChan
			watchResp  clientv3.WatchResponse
			watchEvent *clientv3.Event
			job        *common.Job
			jobName    string
			jobEvent   *common.JobEvent
		)
		watchChan = jobMgr.watcher.Watch(context.TODO(), common.JOB_KILLER_DIR, clientv3.WithPrevKV())
		for watchResp = range watchChan {
			for _, watchEvent = range watchResp.Events {
				switch watchEvent.Type {
				case mvccpb.PUT: //说明这里是任务强杀事件
					jobName = common.ExtractKillerName(string(watchEvent.Kv.Key))
					job = &common.Job{Name: jobName}
					jobEvent = common.BuildJobEvent(common.JOB_EVENT_KILL, job)
					G_Scheduler.PushJobEvent(jobEvent)
				case mvccpb.DELETE: // killer标记过期, 被自动删除
				}
			}
		}
	}()
}

//创建任务执行琐
func (jobMgr *JobMgr) CreateJobLock(jobName string) (jobLock *JobLock) {
	jobLock = InitJobLock(jobMgr.kv, jobMgr.lease, jobName)
	return
}
