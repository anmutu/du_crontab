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

	//初始化任务管理器，一开始这里就启动了监听
	if err = worker.InitJobMgr(); err != nil {
		goto ERR
	}

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
	flag.StringVar(&confFile, "conf", "master.json", "指定master.json")
	flag.Parse()
}
