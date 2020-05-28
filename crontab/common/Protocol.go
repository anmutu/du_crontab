/*
  author='du'
  date='2020/5/26 9:07'
*/
package common

import (
	"encoding/json"
	"strings"
)

//定时任务,要带上"",不然前端还是会取大写的Name
type Job struct {
	Name     string `json:"name"`     //任务名称
	Commond  string `json:"command"`  //shell命令
	CronExpr string `json:"cronExpr"` //cron表达式
}

//返回信息
type Response struct {
	Erron int         `json:"errno"`
	Msg   string      `json:"msg"`
	Data  interface{} `json:"data"`
}

//生成返回信息的函数
func BuildResponse(errno int, msg string, data interface{}) (resp []byte, err error) {
	var (
		response Response
	)
	response.Erron = errno
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

//从jobKey中拿到任务名
func ExtractJobName(jobKey string) string {
	return strings.TrimPrefix(jobKey, JOB_SAVE_DIR)
}
