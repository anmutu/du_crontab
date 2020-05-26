/*
  author='du'
  date='2020/5/24 23:42'
*/
package main

import (
	"fmt"
	"os/exec"
)

func main() {

	cmd := exec.Command("C:\\cygwin64\\bin\\bash.exe", "-c", "ls -al")

	//执行命令，然后捕获子进行的输出
	if output, err := cmd.CombinedOutput(); err != nil {
		fmt.Print(err)
	} else {
		fmt.Print(output)
	}

}
