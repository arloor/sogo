package main

import (
	"encoding/json"
	"github.com/arloor/proxygo/util"
	"github.com/arloor/sogo/utils"
	"log"
	"os"
	"strconv"
)

var localAddr string
var proxyAddr string

func init() {
	log.SetFlags(log.Lshortfile|log.Flags())
	log.SetOutput(os.Stdout)
	log.Println("！！！请务必在行前将proxy.json和pac.txt放置到", utils.GetWorkDir(), "路径下")
	configinit()
	log.Println("配置信息为：", config)
	localAddr = ":" + strconv.Itoa(config.ClientPort)
	proxyAddr = config.ProxyAddr + ":" + strconv.Itoa(config.ProxyPort)
}


type Info struct {
	ProxyAddr  string
	ProxyPort  int
	ClientPort int  //8081，请不要修改
	Relay      bool //如果设为true ，则只做转发，不做加解密
}

var config = Info{
	"proxy",
	8080,
	8888, //8081，请不要修改
	false,
}

func (configInfo Info) String() string {
	str, _ := configInfo.ToJSONString()
	return str
}

//implement JSONObject
func (configInfo Info) ToJSONString() (str string, error error) {
	b, err := json.Marshal(configInfo)
	if err != nil {
		return "", err
	} else {
		return string(b), nil
	}
}

func configinit() {
	configFile, err := os.Open(util.GetWorkDir() + "proxy.json")
	if err != nil {
		log.Println("Error", "打开proxy.json失败，使用默认配置", err)
		return
	}
	bufSize := 1024
	buf := make([]byte, bufSize)
	for {
		total := 0
		n, err := configFile.Read(buf)
		total += n
		if err != nil {
			log.Println("Error", "读取proxy.json失败，使用默认配置", err)
			return
		} else if n < bufSize {
			log.Println("OK", "读取proxy.json成功")
			buf = buf[:total]
			break
		}

	}
	err = json.Unmarshal(buf, &config)
	if err != nil {
		log.Println("Error", "读取proxy.json失败，使用默认配置", err)
		return
	}

}

