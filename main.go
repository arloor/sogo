package main

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/arloor/sogo/mio"
	"log"
	"net"
	"net/http"
	"strings"

	"strconv"
)

var hand = []byte{0x05, 0x02}
var ack = []byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x10, 0x10}

func handler(w http.ResponseWriter, r *http.Request) {
	//写回请求体本身
	bufio.NewReader(r.Body).WriteTo(w)

}
func server8080() {
	http.HandleFunc("/", handler)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Println("serve过程中出错", err)
	}
}

func main() {
	//go server8080()

	//==================

	ln, err := net.Listen("tcp", localAddr)
	if err != nil {
		fmt.Println("监听", localAddr, "失败 ", err)
		return
	}
	defer ln.Close()
	fmt.Println("成功监听 ", ln.Addr())
	for {
		c, err := ln.Accept()
		if err != nil {
			log.Println("接受连接失败 ", err)
		} else {
			go handleClientConnnection(c)
		}
	}
}

func handleClientConnnection(clientCon net.Conn) {
	defer clientCon.Close()
	handshakeErr := handshake(clientCon)
	if handshakeErr != nil {
		log.Println("handshakeErr ", handshakeErr)
		return
	} else {
		//log.Println("与客户端握手成功")
		authErr := auth(clientCon)
		if authErr != nil {
			log.Println("认证错误  ", authErr, clientCon.RemoteAddr())
			return
		}
		addr, getTargetErr := getTargetAddr(clientCon)
		if getTargetErr != nil {
			log.Println("getTargetErr ", getTargetErr)
			return
		} else {
			//开始连接到服务器，并传输
			var serverConn, dialErr = net.Dial("tcp", proxyAddr)
			//var serverConn, dialErr = net.Dial("tcp", addr)
			if dialErr != nil {
				log.Println("dialErr ", dialErr)
				clientCon.Close()
				return
			}
			go handleServerConn(serverConn, clientCon)

			buf := make([]byte, 2048)
			for {

				num, readErr := clientCon.Read(buf)
				if readErr != nil {
					log.Print("readErr ", readErr, clientCon.RemoteAddr())
					clientCon.Close()
					serverConn.Close()
					return
				}
				writeErr := mio.WriteAll(serverConn, mio.AppendHttpRequestPrefix(buf[:num], addr, fakeHost, authorization))
				//writeErr := mio.WriteAll(serverConn, buf[:num])
				if writeErr != nil {
					log.Print("writeErr ", writeErr)
					clientCon.Close()
					serverConn.Close()
					return
				}
				buf = buf[0:]
			}
		}
	}

}

//HTTP/1.1 200 OK
//Content-Type: text/plain; charset=utf-8
//Content-Length: 181
//
//HTTP/1.1 304 Not Modified
//Date: Tue, 09 Apr 2019 08:46:15 GMT
//Server: Apache/2.4.6 (CentOS)
//Connection: Keep-Alive
//Keep-Alive: timeout=5, max=100
//ETag: "37cb-58613717be980"

func handleServerConn(serverConn, clientCon net.Conn) {
	defer serverConn.Close()
	redundancy := make([]byte, 0)
	for {
		redundancyRetain, readerr := read(serverConn, clientCon, redundancy)
		redundancy = redundancyRetain
		if readerr != nil {
			log.Println("readerr", readerr)
			break
		}
	}
}

func read(serverConn, clientConn net.Conn, redundancy []byte) (redundancyRetain []byte, readErr error) {
	buf := make([]byte, 8196)
	num := 0
	contentlength := -1
	prefixAll := false
	prefix := make([]byte, 0)
	//redundancy:=make([]byte,0)
	payload := make([]byte, 0)

	for {
		if len(redundancy) != 0 {
			buf = redundancy
			num = len(redundancy)
			redundancy = redundancy[0:0]
		} else {
			num, readErr = serverConn.Read(buf)
			if readErr != nil {
				return redundancy, readErr
			}
		}

		if num <= 0 {
			return nil, errors.New("读到<=0字节，未预期地情况")
		} else {
			if !prefixAll { //追加到前缀
				prefix = append(prefix, buf[:num]...)
				if index := strings.Index(string(prefix), "\r\n\r\n"); index >= 0 {
					if index+4 < len(prefix) {
						payload = append(payload, prefix[index+4:]...)
					}
					prefix = prefix[:index]
					prefixAll = true
					//分析头部
					headrs := strings.Split(string(prefix), "\r\n")

					requestline := headrs[0]
					parts := strings.Split(requestline, " ")
					if len(parts) < 3 {
						fmt.Println(requestline)
						return nil, errors.New("不是以HTTP/1.1 200 OK这种开头，说明上个响应有问题。")
					}
					//version := parts[0]
					//code := parts[1]
					//msg := parts[2]

					var headmap = make(map[string]string)
					for i := 1; i < len(headrs); i++ {
						headsplit := strings.Split(headrs[i], ": ")
						headmap[headsplit[0]] = headsplit[1]
					}
					if headmap["Content-Length"] == "" {
						contentlength = 0
					} else {
						contentlength, _ = strconv.Atoi(headmap["Content-Length"])
					}
				}
			} else { //追加到payload
				payload = append(payload, buf[:num]...)
			}
		}
		buf = buf[0:]
		if contentlength != -1 && contentlength < len(payload) { //这说明读多了，要把多的放到redundancy
			redundancy = append(redundancy, payload[contentlength:]...)
			payload = payload[:contentlength]
		}
		if contentlength == len(payload) {
			//写会
			mio.Simple(&payload, len(payload))
			writeErr := mio.WriteAll(clientConn, payload)
			if writeErr != nil {
				clientConn.Close()
			}
			return redundancy, writeErr
		}
	}
}

func getTargetAddr(clientCon net.Conn) (string, error) {
	var buf = make([]byte, 1024)
	numRead, err := clientCon.Read(buf)
	if err != nil {
		return "", err
	} else if numRead > 3 && buf[0] == 0X05 && buf[1] == 0X01 && buf[2] == 0X00 {
		if buf[3] == 3 {
			log.Printf("源地址%s 目标域名:%s 目标端口:%s", clientCon.RemoteAddr(), buf[5:5+buf[4]], strconv.Itoa(int(binary.BigEndian.Uint16(buf[5+buf[4]:7+buf[4]]))))
			writeErr := mio.WriteAll(clientCon, ack)
			return string(buf[5:5+buf[4]]) + ":" + strconv.Itoa(int(binary.BigEndian.Uint16(buf[5+buf[4]:7+buf[4]]))), writeErr
		} else if buf[3] == 1 {
			log.Printf("源地址%s 目标域名:%s 目标端口:%s", clientCon.RemoteAddr(), net.IPv4(buf[4], buf[5], buf[6], buf[7]).String(), strconv.Itoa(int(binary.BigEndian.Uint16(buf[8:10]))))
			writeErr := mio.WriteAll(clientCon, ack)
			return net.IPv4(buf[4], buf[5], buf[6], buf[7]).String() + ":" + strconv.Itoa(int(binary.BigEndian.Uint16(buf[8:10]))), writeErr
		} else {
			return "", errors.New("不能处理ipv6")
		}

	} else {
		return "", errors.New("不能处理非CONNECT请求")
	}
}

//读 5 1 0 写回 5 0
func handshake(clientCon net.Conn) error {
	var buf = make([]byte, 300)
	numRead, err := clientCon.Read(buf)
	if err != nil {
		return err
	} else if numRead == 3 && buf[0] == 0X05 && buf[1] == 0X01 && buf[2] == 0X00 {
		return mio.WriteAll(clientCon, hand)
	} else {
		//log.Printf("%d", buf[:numRead])
		return mio.WriteAll(clientCon, hand)
	}
}

func auth(clientCon net.Conn) error {
	var buf = make([]byte, 300)
	numRead, err := clientCon.Read(buf)
	if err != nil {
		return err
	} else if numRead != 3+len(Config.Pass)+len(Config.User) {
		mio.WriteAll(clientCon, []byte{0x01, 0x01})
		return errors.New("密码位数不对")
	} else {
		//log.Printf("%d", buf[:numRead])
		//log.Printf("%d", buf[1])
		if 3+int(buf[1]) > numRead {
			mio.WriteAll(clientCon, []byte{0x01, 0x01})
			return errors.New("用户名密码错误")
		}
		if Config.User == string(buf[2:2+int(buf[1])]) && Config.Pass == string(buf[3+int(buf[1]):numRead]) {
			return mio.WriteAll(clientCon, []byte{0x01, 0x00})
		} else {
			mio.WriteAll(clientCon, []byte{0x01, 0x01})
			return errors.New("用户名密码错误")
		}

	}
	return nil
}
