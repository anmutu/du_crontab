/*
  author='du'
  date='2020/5/26 6:11'
*/
package main

import (
	"du_corntab/crontab/master"
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
	if err = master.InitConfig(confFile); err != nil {
		goto ERR
	}
	fmt.Println("0.1 初始化参数成功:读取json文件，将其赋值给Config的一个对象。")

	//初始化任务管理器,etcd连接与job相关的crud的操作。
	if err = master.InitJobMgr(); err != nil {
		goto ERR
	}
	fmt.Println("0.2 初始化任务管理器成功：由任务管理器进行对job的CRUD的操作。")

	//初始化日志模块，
	if err = master.InitLogMgr(); err != nil {
		goto ERR
	}
	fmt.Println("0.3 初始化mongodb日志模块：可查看任务日志。")

	//启动api服务
	if err = master.InitApiServer(); err != nil {
		goto ERR
	}
	fmt.Println("0.4 初始化api服务成功:由这里的去调用任务管理器具体CRUD的相关函数。")

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
	flag.StringVar(&confFile, "config", "./master.json", "指定master.json")
	flag.Parse()
}
