/*
  author='du'
  date='2020/5/26 9:07'
*/
package common

import (
	"context"
	"encoding/json"
	"github.com/gorhill/cronexpr"
	"strings"
	"time"
)

//定时任务,要带上"",不然前端还是会取大写的Name
type Job struct {
	Name     string `json:"name"`     //任务名称
	Command  string `json:"command"`  //shell命令
	CronExpr string `json:"cronExpr"` //cron表达式
}

//返回信息
type Response struct {
	Errno int         `json:"errno"` //错误代码
	Msg   string      `json:"msg"`
	Data  interface{} `json:"data"`
}

//生成返回信息的函数
func BuildResponse(errno int, msg string, data interface{}) (resp []byte, err error) {
	var (
		response Response
	)
	response.Errno = errno
	response.Msg = msg
	response.Data = data

	resp, err = json.Marshal(response)
	return
}

//反序列化
func UnpackJob(value []byte) (res *Job, err error) {
	job := &Job{}
	if err = json.Unmarshal(value, job); err != nil {
		return
	}
	res = job
	return
}

//从jobKey中拿到任务名,watch模块会用到。
func ExtractJobName(jobKey string) string {
	return strings.TrimPrefix(jobKey, JOB_SAVE_DIR)
}

//变化事件的结构体
type JobEvent struct {
	EventType int //有保存和删除
	Job       *Job
}

//生成JobEvent的函数
func BuildJobEvent(eventType int, job *Job) (jobEvent *JobEvent) {
	return &JobEvent{
		EventType: eventType,
		Job:       job,
	}
}

//任务调度计划,etcd里有什么任务我们就维护跟它一模一样的。
type JobSchedulerPlan struct {
	Job      *Job
	Expr     *cronexpr.Expression //解析好的表达式
	NextTime time.Time            //下次执行时间
}

//构建任务执行计划,参数job,返回JobSchedulerPlan
func BuildJobSchedulerPlan(job *Job) (jobSchedulerPlan *JobSchedulerPlan, err error) {
	var (
		expr *cronexpr.Expression
	)

	//第一步，解析job的cron表达式
	if expr, err = cronexpr.Parse(job.CronExpr); err != nil {
		return
	}

	//第二步，生成任务调度计划对象
	jobSchedulerPlan = &JobSchedulerPlan{
		Job:      job,
		Expr:     expr,
		NextTime: expr.Next(time.Now()),
	}
	return
}

//job执行状态
type JobExecuteInfo struct {
	Job        *Job
	PlanTime   time.Time          //理论调度时间
	RealTime   time.Time          //实际调度时间
	CancelCtx  context.Context    //任务command取消的context,用于强杀任务
	CancelFunc context.CancelFunc //取消任务的函数，用于强杀任务
}

//执行构造状态信息
func BuildJobExecuteInfo(jobScheduler *JobSchedulerPlan) (jobExecuteInfo *JobExecuteInfo) {
	jobExecuteInfo = &JobExecuteInfo{
		Job:      jobScheduler.Job,
		PlanTime: jobScheduler.NextTime,
		RealTime: time.Now(),
	}
	//这里是用于任务强杀。
	jobExecuteInfo.CancelCtx, jobExecuteInfo.CancelFunc = context.WithCancel(context.TODO())
	return
}

//任务执行的结果
type JobExecuteResult struct {
	ExecuteInfo *JobExecuteInfo //执行状态
	Output      []byte          //输出
	Err         error           //执行脚本错误原因
	StartTime   time.Time
	EndTime     time.Time
}

//获取强杀目录里的killName
func ExtractKillerName(killerKey string) (killName string) {
	return strings.TrimPrefix(killerKey, JOB_KILLER_DIR)
}

// 任务执行日志
type JobLog struct {
	JobName      string `json:"jobName" bson:"jobName"`
	Command      string `json:"command" bson:"command"`
	Err          string `json:"err" bson:"err"`
	Output       string `json:"output" bson:"output"`
	PlanTime     int64  `json:"planTime" bson:"planTime"`
	ScheduleTime int64  `json:"scheduleTime" bson:"scheduleTime"`
	StartTime    int64  `json:"startTime" bson:"startTime"`
	EndTime      int64  `json:"endTime" bson:"endTime"`
}
