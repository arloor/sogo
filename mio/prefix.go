package mio

import "strconv"

//POST / HTTP/1.1
//cache-control: no-cache
//Postman-Token: 6b859da0-0a7e-4c4e-a0e2-2544aeef732f
//Content-Type: text/plain
//User-Agent: PostmanRuntime/7.6.1
//Accept: */*
//Host: localhost:8080
//accept-encoding: gzip, deflate
//content-length: 5
//Connection: keep-alive
//
//aaaaa
//
//HTTP/1.1 200 OK
//Date: Tue, 09 Apr 2019 03:26:07 GMT
//Content-Length: 5
//Content-Type: text/plain; charset=utf-8
//Connection: keep-alive
//
//aaaaa

func AppendHttpRequestPrefix(buf []byte, addr string) []byte {
	//simple(&buf,len(buf))
	buf = append([]byte("POST /target?"+addr+" HTTP/1.1\r\nHost: qtgwuehaoisdhuaishdaisuhdasiuhlassjd.com\r\nAccept: */*\r\nContent-Type: text/plain\r\naccept-encoding: gzip, deflate\r\ncontent-length: "+strconv.Itoa(len(buf))+"\r\n\r\n"), buf...)
	return buf
}

//取反
func simple(bufPtr *[]byte, num int) {
	buf := *bufPtr
	for i := 0; i < num; i++ {
		buf[i] = ^buf[i]
	}
}
