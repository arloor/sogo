package main

import (
	"encoding/binary"
	"errors"
	"github.com/arloor/sogo/mio"
	"log"
	"net"
	"strconv"
)

func main()  {
	ln, err := net.Listen("tcp", localAddr)
	if err != nil {
		log.Println("监听", localAddr, "失败 ", err)
		return
	}
	defer ln.Close()
	log.Println("成功监听 ", ln.Addr())
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
	handshakeErr:=handshake(clientCon)
	if handshakeErr!=nil{
		log.Println("handshakeErr ",handshakeErr)
		return
	}else{
		log.Println("与客户端握手成功")
		addr,getTargetErr:=getTargetAddr(clientCon)
		if getTargetErr!=nil{
			log.Println("getTargetErr ",getTargetErr)
			return
		}else {
			//开始连接到服务器，并传输
			var serverConn, dialErr=net.Dial("tcp",addr)
			if dialErr!=nil{
				log.Println("dialErr ",dialErr)
				clientCon.Close()
				return
			}
			go handleServerConn(serverConn,clientCon)

			for{
				buf:=make([]byte,2048)
				num,readErr:=clientCon.Read(buf)
				if readErr!=nil{
					log.Print("readErr ",readErr,clientCon.RemoteAddr())
					clientCon.Close()
					serverConn.Close()
					return
				}
				writeErr:=mio.WriteAll(serverConn,buf[:num])
				if writeErr!=nil{
					log.Print("writeErr ",writeErr)
					clientCon.Close()
					serverConn.Close()
					return
				}
				buf=buf[:0]
			}
		}

	}

}

func handleServerConn(serverConn ,clientCon net.Conn) {

	for{
		buf:=make([]byte,2048)
		num,readErr:=serverConn.Read(buf)
		if readErr!=nil{
			log.Print("readErr ",readErr,serverConn.RemoteAddr())
			clientCon.Close()
			serverConn.Close()
			return
		}
		writeErr:=mio.WriteAll(clientCon,buf[:num])
		if writeErr!=nil{
			log.Print("writeErr ",writeErr)
			clientCon.Close()
			serverConn.Close()
			return
		}
	}
}

func getTargetAddr(clientCon net.Conn) (string, error) {
	var buf = make([]byte, 1024)
	numRead,err:=clientCon.Read(buf)
	if err!=nil{
		return "",err
	}else if numRead>3&&buf[0] == 0X05 && buf[1] == 0X01 && buf[2] == 0X00 {
		if buf[3]!=3{
			return "",errors.New("暂时仅支持目的地址为域名  buf[3]为3请求")
		}
		log.Printf("目的地址类型:%d 域名长度:%d 目标域名:%s 目标端口:%s",buf[3],buf[4],buf[5:5+buf[4]],strconv.Itoa(int(binary.BigEndian.Uint16(buf[5+buf[4]:7+buf[4]]))))
		writeErr:=mio.WriteAll(clientCon,[]byte{0x05,0x00,0x00,0x01,0x00,0x00,0x00,0x00,0x10,0x10})
		return string(buf[5:5+buf[4]])+":"+strconv.Itoa(int(binary.BigEndian.Uint16(buf[5+buf[4]:7+buf[4]]))),writeErr
	}else {
		return "",errors.New("不能处理非CONNECT请求")
	}
}

//读 5 1 0 写回 5 0
func handshake(clientCon net.Conn) error {
	var buf = make([]byte, 300)
	numRead,err:=clientCon.Read(buf)
	if err!=nil{
		return err
	}else if numRead==3&&buf[0] == 0X05 && buf[1] == 0X01 && buf[2] == 0X00 {
		return mio.WriteAll(clientCon,[]byte{0x05,0x00})
	}else {
		return errors.New("不能处理该socks5握手")
	}
}