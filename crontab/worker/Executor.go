/*
  author='du'
  date='2020/5/31 5:07'
*/
package worker

import (
	"context"
	"du_corntab/crontab/common"
	"os/exec"
	"time"
)

type Executor struct {
}

var (
	G_Executor *Executor
)

//初始化Executor
func InitExecutor() (err error) {
	G_Executor = &Executor{}
	return
}

//执行任务
func (executor *Executor) ExecutorJob(info *common.JobExecuteInfo) {
	go func() {
		var (
			cmd    *exec.Cmd
			output []byte
			err    error
			result *common.JobExecuteResult
		)

		//执行shell命令且捕获输出
		result = &common.JobExecuteResult{
			ExecuteInfo: info,
			Output:      make([]byte, 0),
		}
		result.StartTime = time.Now()
		//cmd=exec.CommandContext(context.TODO(),"/bin/bash","-c",info.Job.Commond)
		cmd = exec.CommandContext(context.TODO(), "C:\\cygwin64\\bin\\bash.exe", "-c", info.Job.Commond)
		output, err = cmd.CombinedOutput()
		result.Output = output
		result.Err = err
		result.EndTime = time.Now()

		//任务执行完成后，将结果告诉Scheduler,Scheduler则会从executingTable表中删除掉执行的数据记录
		G_Scheduler.PushJobResult(result)

	}()

}
