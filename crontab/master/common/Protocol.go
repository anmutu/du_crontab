/*
  author='du'
  date='2020/5/26 9:07'
*/
package common

type Job struct {
	Name     string `json:name`     //任务名称
	Commond  string `json:command`  //shell命令
	CronExpr string `json:cronExpr` //cron表达式
}
