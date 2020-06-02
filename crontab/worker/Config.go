/*
  author='du'
  date='2020/5/27 23:43'
*/
package worker

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type Config struct {
	EtcdEndpoints   []string `json:"etcdEndpoints"`
	EtcdDialTimeout int      `json:"etcdDialTimeout"`
	MongodbUri      string   `json:"mongodbUri"`
}

var (
	G_config *Config //单例对象，指针。
)

//初始化配置文件
func InitConfig(filename string) (err error) {
	var (
		content []byte
		conf    Config
	)

	//把配置文件读进来
	if content, err = ioutil.ReadFile(filename); err != nil {
		return
	}
	//json反序列化
	if err = json.Unmarshal(content, &conf); err != nil {
		return
	}
	//赋值给单例config
	G_config = &conf
	fmt.Println("1.初始化参数成功：成功将worker.json解析出来给到单列对象Config。")
	fmt.Println(conf)
	return
}
