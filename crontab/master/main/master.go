/*
  author='du'
  date='2020/5/26 6:11'
*/
package main

import (
	"du_corntab/crontab/master"
	"fmt"
	"runtime"
)

func main() {
	var (
		err error
	)

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
