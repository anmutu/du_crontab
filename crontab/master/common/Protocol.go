/*
  author='du'
  date='2020/5/26 9:07'
*/
package common

import "encoding/json"

//定时任务
type Job struct {
	Name     string `json:name`     //任务名称
	Commond  string `json:command`  //shell命令
	CronExpr string `json:cronExpr` //cron表达式
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
