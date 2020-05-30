/*
  author='du'
  date='2020/5/30 1:46'
*/
package worker

import (
	"du_corntab/crontab/common"
	"fmt"
)

//任务调度结构体
type Scheduler struct {
	jobEventChan chan *common.JobEvent               //etcd任务事件队列。
	jobPlanTable map[string]*common.JobSchedulerPlan //任务设计计划表，key为任务名。
}

var (
	G_Scheduler *Scheduler
)

//初始化调度协程
func InitScheduler(err error) {
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
		jobEvent *common.JobEvent
	)
	for {
		select {
		case jobEvent = <-scheduler.jobEventChan:
			//对内存中维护的任务列表进行CRUD
			scheduler.handleJobEvent(jobEvent)
		}
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
	scheduler.jobEventChan <- jobEvent
}
