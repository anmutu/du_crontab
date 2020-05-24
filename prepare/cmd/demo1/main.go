/*
  author='du'
  date='2020/5/24 23:11'
*/
package main

import "os/exec"


func main(){
	cmd:= exec.Command("C:\\cygwin64\\bin\\bash.exe")
	err:=cmd.Run()
	if err!=nil{
		print(err)
	}else{
		print("success")
	}
}
