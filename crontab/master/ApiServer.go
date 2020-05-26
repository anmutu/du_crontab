/*
  author='du'
  date='2020/5/26 6:11'
*/
package master

import (
	"du_corntab/crontab/master/common"
	"encoding/json"
	"net"
	"net/http"
	"strconv"
	"time"
)

//任务的http接口
type ApiServer struct {
	httpServer *http.Server
}

var (
	//这是一个单例对象
	G_apiServer *ApiServer
)

//初始化服务
func InitApiServer() (err error) {
	var (
		mux        *http.ServeMux
		listener   net.Listener
		httpServer *http.Server
	)
	//路由配置
	mux = http.NewServeMux()
	mux.HandleFunc("/job/save", handleJobSave)
	//启动监听
	listener, err = net.Listen("tcp", ":"+strconv.Itoa(G_config.ApiPort))
	if err != nil {
		return
	}
	//创建一个http服务
	httpServer = &http.Server{
		ReadTimeout:  time.Duration(G_config.ApiReadTimeOut) * time.Millisecond,
		WriteTimeout: time.Duration(G_config.ApiWriteTimeOut) * time.Millisecond,
		Handler:      mux,
	}
	//给单例对象赋值
	G_apiServer = &ApiServer{
		httpServer: httpServer,
	}
	//启动服务端
	go httpServer.Serve(listener)
	return
}

//保存任务接口
//内容是 job={"name":"job1","command":"echo hi","cronExpr":"* * * * *"}
func handleJobSave(resp http.ResponseWriter, req *http.Request) {
	var (
		err       error
		postJob   string
		job       common.Job
		oldJob    *common.Job
		respBytes []byte
	)
	//第一步，解析Post的表单
	if err = req.ParseForm(); err != nil {
		goto ERR
	}

	//表单中取出job字段。
	postJob = req.PostForm.Get("job")
	//反序列化
	if err = json.Unmarshal([]byte(postJob), &job); err != nil {
		goto ERR
	}
	//保存job,调用JobMgr的方法。
	if oldJob, err = G_JobMgr.SaveJob(&job); err != nil {
		goto ERR
	}
	//返回正常消息结构体
	if respBytes, err = common.BuildResponse(0, "success", oldJob); err == nil {
		resp.Write(respBytes)
	}
	return
ERR:
	//返回异常消息结构体
	if respBytes, err = common.BuildResponse(-1, err.Error(), nil); err == nil {
		resp.Write(respBytes)
	}

}
