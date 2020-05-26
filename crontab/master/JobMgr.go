/*
  author='du'
  date='2020/5/26 9:10'
*/
package master

import (
	"context"
	"du_corntab/crontab/master/common"
	"encoding/json"
	"go.etcd.io/etcd/clientv3"
	"time"
)

//任务管理器
type JobMgr struct {
	client *clientv3.Client //这里用的是指针
	kv     clientv3.KV
	lease  clientv3.Lease
}

var (
	G_JobMgr *JobMgr //单例对象,指针
)

//初始化管理器
func InitJogMgr() (err error) {
	var (
		config clientv3.Config
		client *clientv3.Client
		kv     clientv3.KV
		lease  clientv3.Lease
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

	//得到kv和lease租约
	kv = clientv3.NewKV(client)
	lease = clientv3.NewLease(client)

	//将之赋值给单例对像
	G_JobMgr = &JobMgr{
		client: client,
		kv:     kv,
		lease:  lease,
	}
	return
}

//保存job到etcd。
func (JobMgr *JobMgr) SaveJob(job *common.Job) (oldJob *common.Job, err error) {
	var (
		jobKey    string
		jobValue  []byte
		putResp   *clientv3.PutResponse
		oldJobObj common.Job
	)

	//拿到jobKey和jobValue,将交之保存到etcd中。
	jobKey = "/cron/jobs" + job.Name
	if jobValue, err = json.Marshal(job); err != nil {
		return
	}

	if putResp, err = JobMgr.kv.Put(context.TODO(), jobKey, string(jobValue), clientv3.WithPrevKV()); err != nil {
		//如果是更新，就返回以前的值，如果反序列化不成功，则将对象和err都置为nil。
		if putResp.PrevKv.Value != nil {
			if err = json.Unmarshal(putResp.PrevKv.Value, &oldJobObj); err != nil {
				err = nil
				oldJob = nil
			} else {
				oldJob = &oldJobObj
			}
		}
		return nil, err
	}
	return
}
