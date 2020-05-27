/*
  author='du'
  date='2020/5/26 6:11'
*/
package master

import (
	"du_corntab/crontab/master/common"
	"encoding/json"
	"fmt"
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

//初始化服务,路由配置等。
func InitApiServer() (err error) {
	var (
		mux        *http.ServeMux
		listener   net.Listener
		httpServer *http.Server
	)

	//路由配置
	mux = http.NewServeMux()
	mux.HandleFunc("/job/save", handleJobSave)
	mux.HandleFunc("/job/delete", handleJobDelete)
	mux.HandleFunc("/job/list", handleJobList)
	mux.HandleFunc("/job/killjob", handleJobKill)

	// 静态文件目录
	staticDir := http.Dir(G_config.Web)
	staticHandler := http.FileServer(staticDir)
	mux.Handle("/", http.StripPrefix("/", staticHandler))

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
	fmt.Println("初始化服务成功。")
	return
}

//获取所有crontab列表任务
func handleJobList(resp http.ResponseWriter, req *http.Request) {
	var (
		err       error
		jobList   []*common.Job
		respBytes []byte
	)

	//调用JobMgr里的函数
	if jobList, err = G_JobMgr.ListJobs(); err != nil {
		//goto ERR
		if respBytes, err = common.BuildResponse(-1, "jobMgr里获取失败", err.Error()); err == nil {
			resp.Write(respBytes)
		}
	}

	//正常返回
	if respBytes, err = common.BuildResponse(0, "success", jobList); err == nil {
		resp.Write(respBytes)
	}
	return

	//ERR:
	//	if respBytes, err = common.BuildResponse(-1, err.Error(), nil); err == nil {
	//		resp.Write(respBytes)
	//	}

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
		//goto ERR
		respBytes, err = common.BuildResponse(-1, "get form failed", err.Error())
		resp.Write(respBytes)
	}

	//表单中取出job字段。
	postJob = req.PostForm.Get("job")
	//反序列化
	if err = json.Unmarshal([]byte(postJob), &job); err != nil {
		//goto ERR
		respBytes, err = common.BuildResponse(-2, "json unmarshal failed", err.Error())
		resp.Write(respBytes)
	}
	//保存job,调用JobMgr的方法。
	if oldJob, err = G_JobMgr.SaveJob(&job); err != nil {
		//goto ERR
		respBytes, err = common.BuildResponse(-3, "save job 2 etcd failed", err.Error())
		resp.Write(respBytes)
	}
	//返回正常消息结构体
	if respBytes, err = common.BuildResponse(0, "success", oldJob); err == nil {
		resp.Write(respBytes)
	}
	return
	//ERR:
	//	//返回异常消息结构体
	//	if respBytes, err = common.BuildResponse(-1, err.Error(), nil); err == nil {
	//		resp.Write(respBytes)
	//	}

}

//删除任务接口
//传入name
func handleJobDelete(resp http.ResponseWriter, req *http.Request) {
	var (
		err       error
		name      string
		oldJob    *common.Job
		respBytes []byte
	)
	if err = req.ParseForm(); err != nil {
		goto ERR
	}
	name = req.PostForm.Get("name")
	//调用JobMgr里的函数删除etcd里的Job
	if oldJob, err = G_JobMgr.DeleteJob(name); err != nil {
		goto ERR
	}
	//正常返回
	if respBytes, err = common.BuildResponse(0, "success", oldJob); err == nil {
		resp.Write(respBytes)
	}
	return

ERR:
	if respBytes, err = common.BuildResponse(-1, err.Error(), nil); err == nil {
		resp.Write(respBytes)
	}
}

//强杀任务
//传入 /job/kill name=job
func handleJobKill(resp http.ResponseWriter, req *http.Request) {
	var (
		err       error
		name      string
		respBytes []byte
	)
	if err = req.ParseForm(); err != nil {
		goto ERR
	}

	name = req.PostForm.Get("name")

	//调用JobMgr里的函数
	if err = G_JobMgr.KillJob(name); err != nil {
		return
	}

	//正常返回
	if respBytes, err = common.BuildResponse(0, "success", nil); err == nil {
		resp.Write(respBytes)
	}
	return

ERR:
	if respBytes, err = common.BuildResponse(-1, err.Error(), nil); err == nil {
		resp.Write(respBytes)
	}

}
