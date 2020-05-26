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

	//初始化配置,这里的confFile就是通过initArgs得到的。
	if err = master.InitConfig(confFile); err != nil {
		goto ERR
	}

	//初始化线程
	initEnv()

	//启动api服务
	if err = master.InitApiServer(); err != nil {
		goto ERR
	}

	return

ERR:
	fmt.Println(err)
}

func initEnv() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

//解析命令行参数，最后将之放到配置文件路径中
func initArgs() {
	flag.StringVar(&confFile, "conf", "master.json", "指定master.json")
	flag.Parse()
}
