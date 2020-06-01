/*
  author='du'
  date='2020/5/30 1:46'
*/
package worker

import (
	"du_corntab/crontab/common"
	"fmt"
	"time"
)

//任务调度结构体
type Scheduler struct {
	jobEventChan      chan *common.JobEvent               //etcd任务事件队列。
	jobPlanTable      map[string]*common.JobSchedulerPlan //任务调度计划表，key为任务名。
	jobExecutingTable map[string]*common.JobExecuteInfo   //任务执行表key为任务名。
	jobResultChan     chan *common.JobExecuteResult       //任务执行后结果的一个队列。
}

var (
	G_Scheduler *Scheduler
)

//初始化调度协程
func InitScheduler() (err error) {
	G_Scheduler = &Scheduler{
		jobEventChan:      make(chan *common.JobEvent, 1000),
		jobPlanTable:      make(map[string]*common.JobSchedulerPlan),
		jobExecutingTable: make(map[string]*common.JobExecuteInfo),
		jobResultChan:     make(chan *common.JobExecuteResult, 1000),
	}

	//启动调度协程
	go G_Scheduler.schedulerLoop()
	return
}

//检测。for循环检测所有的任务
func (scheduler *Scheduler) schedulerLoop() {
	var (
		jobEvent       *common.JobEvent
		schedulerAfter time.Duration
		schedulerTimer *time.Timer
		jobResult      *common.JobExecuteResult
	)

	//初始化一次
	schedulerAfter = scheduler.TryScheduler()
	schedulerTimer = time.NewTimer(schedulerAfter)

	for {
		select {
		case jobEvent = <-scheduler.jobEventChan:
			//对内存中维护的任务列表jobPlanTable进行与etcd里的job进行同步操作。
			scheduler.handleJobEvent(jobEvent)
		case <-schedulerTimer.C: //说明最近的任务到期了
		case jobResult = <-scheduler.jobResultChan:
			scheduler.handleJobResult(jobResult)
		}
		//调度任务
		schedulerAfter = scheduler.TryScheduler()
		schedulerTimer.Reset(schedulerAfter)
	}
}

//处理任务的事件
//也就是往Scheduler里的jobPlanTable维护一个跟etcd里一模一样的job。
func (scheduler Scheduler) handleJobEvent(jobEvent *common.JobEvent) {
	var (
		jobScheduler *common.JobSchedulerPlan
		err          error
		jobExist     bool
	)
	switch jobEvent.EventType {
	case common.JOB_EVENT_SAVE:
		if jobScheduler, err = common.BuildJobSchedulerPlan(jobEvent.Job); err != nil {
			return
		}
		scheduler.jobPlanTable[jobEvent.Job.Name] = jobScheduler
		fmt.Println("向Scheduler里的jobPlanTable表里做维护操作=>scheduler里检测到有增加任务:", scheduler.jobPlanTable[jobEvent.Job.Name].Job.Name)
	case common.JOB_EVENT_DELETE:
		if jobScheduler, jobExist = scheduler.jobPlanTable[jobEvent.Job.Name]; jobExist {
			delete(scheduler.jobPlanTable, jobEvent.Job.Name)
			fmt.Println("向Scheduler里的jobPlanTable表里做维护操作=>scheduler里检测到有刪除任务:", scheduler.jobPlanTable[jobEvent.Job.Name].Job.Name)
		}
	}
}

//推送任务变化事件,就是把jobEvent任务的到Scheduler里的jobEventChan里。
//JobMgr那边检测到任务变化则会调用此函数。
func (scheduler *Scheduler) PushJobEvent(jobEvent *common.JobEvent) {
	if scheduler == nil {
		scheduler = &Scheduler{}
	}
	scheduler.jobEventChan <- jobEvent
	fmt.Println("成功将数据推送到scheduler.jobEventChan，其任务名称为", jobEvent.Job.Name)
}

//重新计算任务调度状态
func (scheduler *Scheduler) TryScheduler() (schedulerAfter time.Duration) {
	//遍历所有任务。过期的任务立即执行。统计最近要过期的任务。
	var (
		jobSchedulerPlan *common.JobSchedulerPlan
		now              time.Time
		nearTime         *time.Time
	)
	//极端情况,没有需要执行的job
	if len(scheduler.jobPlanTable) == 0 {
		schedulerAfter = 1 * time.Second
		return
	}

	//遍历scheduler肚子里的任务调度计划表这个map
	now = time.Now()
	for _, jobSchedulerPlan = range scheduler.jobPlanTable {
		if jobSchedulerPlan.NextTime.Before(now) || jobSchedulerPlan.NextTime.Equal(now) {
			fmt.Println("scheduler,执行任务:", jobSchedulerPlan.Job.Name)
			//尝试执行任务
			scheduler.TryStartJob(jobSchedulerPlan)
			jobSchedulerPlan.NextTime = jobSchedulerPlan.Expr.Next(now) //更新下次时间
		}
		//统计最近一个任务要过期的时间
		if nearTime == nil || jobSchedulerPlan.NextTime.Before(*nearTime) {
			nearTime = &jobSchedulerPlan.NextTime
		}
	}
	schedulerAfter = (*nearTime).Sub(now)
	return
}

//尝试开始执行任务
//如果任务正在执行，那么跳过本次任务。
func (scheduler *Scheduler) TryStartJob(jobPlan *common.JobSchedulerPlan) {
	var (
		jobExecuteInfo *common.JobExecuteInfo
		jobExecuting   bool //是否正在执行。也就是看执行表里是否有。
	)
	//正在执行则返回，不用管
	if jobExecuteInfo, jobExecuting = scheduler.jobExecutingTable[jobPlan.Job.Name]; jobExecuting {
		fmt.Println("scheduler trystartjob 任务正在执行，尚未退出。跳过执行。")
		return
	}
	jobExecuteInfo = common.BuildJobExecuteInfo(jobPlan)           //创建执行状态的相差信息
	scheduler.jobExecutingTable[jobPlan.Job.Name] = jobExecuteInfo //保存执行状态
	//TODO 执行任务
	G_Executor.ExecutorJob(jobExecuteInfo)
	fmt.Println("scheduler trystartjob "+jobExecuteInfo.Job.Name, jobExecuteInfo.PlanTime, jobExecuteInfo.RealTime)
}

//回传任务结果到channel的函数。
func (scheduler *Scheduler) PushJobResult(jobResult *common.JobExecuteResult) {
	scheduler.jobResultChan <- jobResult
}

//处理由executor传过来的结果
//把这个执行过的job从jobExecutingTable表中删除掉
func (scheduler *Scheduler) handleJobResult(result *common.JobExecuteResult) {
	delete(scheduler.jobExecutingTable, result.ExecuteInfo.Job.Name)
	fmt.Println("scheduler.handleJobResult收到结果:", result.ExecuteInfo.Job.Name, string(result.Output), result.Err)
}
