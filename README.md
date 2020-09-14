# du_crontab
http://zmfei4.com:8070/  

### 项目介绍
这个项目可以认为是一个升级版本的crontab。

传统的方案会有如下问题：
* 配置任务时，需要ssh登录到服务器上去然后进行操作。
* 服务器宕机，任务将终止任务调度，需要人工迁移。
* 排查任务低效，不方便查看任务的状态与错误的输出。

那么现在的分布式任务调度需要做到了几点：
* 可视化的web后台，方便任务进行管理。
* 追踪任务执行状态，采集任务输出，可视化log查看。

### 目录介绍
* common 这里放一些公共的结构体或者公用函数等。
* master 为master节点，提供将job信息crud到etcd的api和将日志信息cr到mongodb的api。
* worker 负责监控job,并将监控到的到job去执行且日志写入到mongodb。

### 环境
```
Mongdb4.0
Etcd v.3.3.8
生产部署环境：centos 7.1。
```

### 如何run起来
```
windows用户如果在本机跑的话需要先下载cygwin，将之安装，然后到worker下的Executor.go里修改执行代码，指向安装cygin位置:
cmd = exec.CommandContext(info.CancelCtx, "C:\\cygwin64\\bin\\bash.exe", "-c", info.Job.Command)
linux用户则是
cmd=exec.CommandContext(info.CancelCtx,"/bin/bash","-c",info.Job.Command)
master.json和worker.json里的列表集群可以换成你自己的，也可以配置多个，避免单点故障。
run master和run worker.
http://localhost:8070/就可以去管理job了。
```

### 可以做进一步的工作
```
1. 界面逻辑优化，再次确认等。
2. 任务失败告警。
3. 健康节点检测。
4. 生产环境出现的问题。
```




