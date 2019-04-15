package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/arloor/sogo/utils"
	"log"
	"os"
	"strconv"
	"sync"
)

var configFilePath string = utils.GetWorkDir() + "sogo.json" //绝对路径或相对路径

var localAddr string
var proxyAddr string
var authorization string
var prefix1 string = "POST /target?at="
var prefix2 string
var prefix3 = "\r\n\r\n"

//prefix= prefix1+target+prefix2+length+prefix

var pool = &sync.Pool{
	New: func() interface{} {
		log.Println("new 1")
		return make([]byte, 9192)
	},
}

var pool2 = &sync.Pool{
	New: func() interface{} {
		log.Println("new 22222222")
		return make([]byte, 9192)
	},
}

const fakeHost string = "qtgwuehaoisdhuaishdaisuhdasiuhlassjd.com"

func printUsage() {
	fmt.Println("运行方式： sogo [-c  configFilePath ]  若不使用 -c指定配置文件，则默认使用" + configFilePath)
}

func init() {

	printUsage()

	if len(os.Args) == 3 && os.Args[1] == "-c" {
		configFilePath = os.Args[2]
	}

	log.SetOutput(os.Stdout)
	log.SetFlags(log.Lshortfile | log.Flags())
	configinit()
	log.Println("配置信息为：", Config)
	server := Config.Servers[Config.Use]
	setServerConfig(server)
	if !Config.Dev {
		log.Println("已启动sogo客户端，请在sogo_" + strconv.Itoa(Config.ClientPort) + ".log查看详细日志")
		f, _ := os.OpenFile("sogo_"+strconv.Itoa(Config.ClientPort)+".log", os.O_WRONLY|os.O_CREATE|os.O_SYNC|os.O_APPEND, 0755)
		log.SetOutput(f)
	}

	localAddr = ":" + strconv.Itoa(Config.ClientPort)

}

//设置三个参数
func setServerConfig(server Server) {
	proxyAddr = server.ProxyAddr + ":" + strconv.Itoa(server.ProxyPort)
	authorization = base64.StdEncoding.EncodeToString([]byte(server.UserName + ":" + server.Password))
	prefix2 = " HTTP/1.1\r\nHost: " + fakeHost + "\r\nAuthorization: Basic " + authorization + "\r\nAccept: */*\r\nContent-Type: text/plain\r\naccept-encoding: gzip, deflate\r\ncontent-length: "
	log.Println("服务器配置：", proxyAddr, "认证信息：", "Basic "+authorization)
}

type Server struct {
	ProxyAddr string
	ProxyPort int
	UserName  string
	Password  string
}

type Info struct {
	Dev        bool
	ClientPort int
	Use        int
	Servers    []Server //8081，请不要修改
}

var Config = Info{
	true,
	7777,
	0,
	[]Server{
		Server{"proxy", 80, "a", "b"},
	},
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
	configFile, err := os.Open(configFilePath)
	defer configFile.Close()
	if err != nil {
		log.Println("Error", "打开"+configFilePath+"失败，使用默认配置", err)
		return
	}
	bufSize := 1024
	buf := make([]byte, bufSize)
	for {
		total := 0
		n, err := configFile.Read(buf)
		total += n
		if err != nil {
			log.Println("Error", "读取"+configFilePath+"失败，使用默认配置", err)
			return
		} else if n < bufSize {
			log.Println("OK", "读取"+configFilePath+"成功")
			buf = buf[:total]
			break
		}

	}
	err = json.Unmarshal(buf, &Config)
	if err != nil {
		log.Println("Error", "读取"+configFilePath+"失败，使用默认配置", err)
		return
	}

}
