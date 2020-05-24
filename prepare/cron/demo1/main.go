/*
  author='du'
  date='2020/5/25 0:52'
*/

package main

import (
	"fmt"
	"github.com/gorhill/cronexpr"
	"time"
)

func main() {
	//每5秒钟执行一次。cronexpr支持的粒度有秒和年。
	expr, err := cronexpr.Parse("*/5 * * * * * *")
	if err != nil {
		fmt.Print(err)
		return
	}

	nowTime := time.Now()
	nextTime := expr.Next(nowTime)
	fmt.Printf("现在的时间是：%s,下次执行的时间是：%s", nowTime, nextTime)

}
