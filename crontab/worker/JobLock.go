/*
  author='du'
  date='2020/5/31 8:24'
*/
package worker

import (
	"context"
	"du_corntab/crontab/common"
	"github.com/coreos/etcd/clientv3"
)

//import "go.etcd.io/etcd/clientv3"

type JobLock struct {
	kv         clientv3.KV
	lease      clientv3.Lease
	jobName    string
	leaseId    clientv3.LeaseID
	isLock     bool
	cancelFunc context.CancelFunc // 用于终止自动续租
}

//初始化锁
func InitJobLock(kv clientv3.KV, lease clientv3.Lease, jobName string) (jobLock *JobLock) {
	jobLock = &JobLock{
		kv:      kv,
		lease:   lease,
		jobName: jobName,
	}
	return
}

//尝试上锁的函数。
//创建租约。自动续租。创建事务。事务抢锁。成功返回，若失败则释放租约。
func (jobLock *JobLock) TryLock() (err error) {
	var (
		leaseGrantResp *clientv3.LeaseGrantResponse
		cancelCtx      context.Context
		cancelFunc     context.CancelFunc
		keepRespChan   <-chan *clientv3.LeaseKeepAliveResponse //只读的channel
		txn            clientv3.Txn
		lockKey        string
		txnResp        *clientv3.TxnResponse
	)

	//第一步，创建租约。
	if leaseGrantResp, err = jobLock.lease.Grant(context.TODO(), 5); err != nil {
		return
	}

	//第二步，续租。续租失败就等于说是上锁失败了。
	cancelCtx, cancelFunc = context.WithCancel(context.TODO())
	if keepRespChan, err = jobLock.lease.KeepAlive(cancelCtx, leaseGrantResp.ID); err != nil {
		cancelFunc()                                            //取消自动续租
		jobLock.lease.Revoke(context.TODO(), leaseGrantResp.ID) //释放租约
		return
	}

	//第三步，处理续租的应答。
	go func() {
		var (
			keepResp *clientv3.LeaseKeepAliveResponse
		)
		for {
			select {
			case keepResp = <-keepRespChan:
				if keepResp == nil {
					goto END
				}
			}
		}
	END:
	}()

	//第四步，创建事务txn并且抢锁和提交事务
	txn = jobLock.kv.Txn(context.TODO())
	lockKey = common.JOB_LOCK_DIR + jobLock.jobName
	txn.If(clientv3.Compare(clientv3.CreateRevision(lockKey), "=", 0)).
		Then(clientv3.OpPut(lockKey, "", clientv3.WithLease(leaseGrantResp.ID))).
		Else(clientv3.OpGet(lockKey))
	if txnResp, err = txn.Commit(); err != nil {
		cancelFunc()                                            //取消自动续租
		jobLock.lease.Revoke(context.TODO(), leaseGrantResp.ID) //释放租约
		return
	}

	//第五步，返回数据
	if !txnResp.Succeeded {
		err = common.ERR_LOCK_ALREADY_OCCUPIED
		cancelFunc()                                            //取消自动续租
		jobLock.lease.Revoke(context.TODO(), leaseGrantResp.ID) //释放租约
		return
	}
	jobLock.leaseId = leaseGrantResp.ID
	jobLock.isLock = true
	jobLock.cancelFunc = cancelFunc
	return
}

func (jobLock *JobLock) UnLock() (err error) {
	if jobLock.isLock {
		jobLock.cancelFunc()                                  //取消自动续约协程
		jobLock.lease.Revoke(context.TODO(), jobLock.leaseId) //释放租约
	}
	return
}
