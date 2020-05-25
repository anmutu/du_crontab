/*
  author='du'
  date='2020/5/26 6:11'
*/
package master

import (
	"net"
	"net/http"
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
	listener, err = net.Listen("tcp", ":8070")
	if err != nil {
		return
	}
	//创建一个http服务
	httpServer = &http.Server{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
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
func handleJobSave(w http.ResponseWriter, r *http.Request) {

}
