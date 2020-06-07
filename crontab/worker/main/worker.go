/*
  author='du'
  date='2020/5/27 23:40'
*/
package main

import (
	"du_corntab/crontab/worker"
	"flag"
	"fmt"
	"runtime"
	"time"
)

var (
	confFile string
)

func main() {
	var (
		err error
	)

	//初始化命令行参数
	initArgs()

	//初始化线程
	initEnv()

	//初始化配置,这里的confFile就是通过initArgs得到的。
	if err = worker.InitConfig(confFile); err != nil {
		goto ERR
	}
	fmt.Println("0.1 初始化参数成功：成功将worker.json解析出来给到单列对象Config。")

	//初始化日志协程
	if err = worker.InitLogSink(); err != nil {
		//goto ERR
		fmt.Println("InitLogSink:", err)

	}
	fmt.Println("0.2 初始化日志模块成功。")

	//启动执行器
	if err = worker.InitExecutor(); err != nil {
		goto ERR
	}
	fmt.Println("0.3 启动执行器exector成功。")

	//初始化调度器
	if err = worker.InitScheduler(); err != nil {
		goto ERR
	}
	fmt.Println("0.4 初始化调度器scheduler成功。")

	//初始化任务管理器，一开始这里就启动了监听，就会一直把有变化的job发送到scheduler的channel中。
	if err = worker.InitJobMgr(); err != nil {
		goto ERR
	}
	fmt.Println("0.5 初始化任务管理器成功。即将监听现有任务，任务变化（修改和删除），任务强杀。")

	//让程序常驻
	for {
		time.Sleep(1 * time.Second)
	}

	return

ERR:
	fmt.Println(err)
}

//初始化线程
func initEnv() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

//解析命令行参数，最后将之放到配置文件路径中
func initArgs() {
	//flag.StringVar(&confFile, "conf", "worker.json", "worker.json")
	//flag.Parse()
	flag.StringVar(&confFile, "config", "./worker.json", "worker.json")
	flag.Parse()
}
