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
	jobEventChan chan *common.JobEvent               //etcd任务事件队列。
	jobPlanTable map[string]*common.JobSchedulerPlan //任务调度计划表，key为任务名。
}

var (
	G_Scheduler *Scheduler
)

//初始化调度协程
func InitScheduler() (err error) {
	G_Scheduler = &Scheduler{
		jobEventChan: make(chan *common.JobEvent, 1000),
		jobPlanTable: map[string]*common.JobSchedulerPlan{},
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
	)

	//初始化一次
	schedulerAfter = scheduler.TryScheduler()
	schedulerTimer = time.NewTimer(schedulerAfter)

	for {
		select {
		case jobEvent = <-scheduler.jobEventChan:
			//对内存中维护的任务列表进行CRUD
			scheduler.handleJobEvent(jobEvent)
		case <-schedulerTimer.C:

		}
		//调度任务
		schedulerAfter = scheduler.TryScheduler()
		schedulerTimer.Reset(schedulerAfter)
	}
}

//处理任务的事件
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
		fmt.Printf("scheduler里检测到有增加任务:%v", scheduler.jobPlanTable[jobEvent.Job.Name])

	case common.JOB_EVENT_DELETE:
		if jobScheduler, jobExist = scheduler.jobPlanTable[jobEvent.Job.Name]; jobExist {
			delete(scheduler.jobPlanTable, jobEvent.Job.Name)
			fmt.Printf("scheduler里检测到有刪除任务:%v", scheduler.jobPlanTable[jobEvent.Job.Name])
		}
	}
}

//推送任务变化事件,就是把jobEvent任务的到Scheduler里的jobEventChan里。
func (scheduler *Scheduler) PushJobEvent(jobEvent *common.JobEvent) {
	if scheduler == nil {
		scheduler = &Scheduler{}
	}
	scheduler.jobEventChan <- jobEvent
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
