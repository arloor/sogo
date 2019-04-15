package mio

import (
	"encoding/base64"
	"strconv"
)

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

func AppendHttpRequestPrefix(buf []byte, targetAddr string, prefix1 string, prefix2 string, prefix3 string) []byte {
	content := buf[1000:]
	prefix := buf[:1000]
	Simple(&content, len(content))
	contentLength := strconv.Itoa(len(content))
	// 演示base64编码
	addrBase64 := base64.NewEncoding("abcdefghijpqrzABCKLMNOkDEFGHIJl345678mnoPQRSTUVstuvwxyWXYZ0129+/").EncodeToString([]byte(targetAddr))
	prefixEnd := 1000
	prefixLength := len(prefix1) + len(addrBase64) + len(prefix2) + len(contentLength) + len(prefix3)
	prefixStart := prefixEnd - prefixLength
	prefixIndex := prefixStart
	for _, x := range prefix1 {
		prefix[prefixIndex] = byte(x)
		prefixIndex++
	}
	for _, x := range addrBase64 {
		prefix[prefixIndex] = byte(x)
		prefixIndex++
	}
	for _, x := range prefix2 {
		prefix[prefixIndex] = byte(x)
		prefixIndex++
	}
	for _, x := range contentLength {
		prefix[prefixIndex] = byte(x)
		prefixIndex++
	}
	for _, x := range prefix3 {
		prefix[prefixIndex] = byte(x)
		prefixIndex++
	}
	buf = buf[prefixStart:]
	//println(len(buf))  lenbuf就是实际上传字节数
	return buf
}

//取反
func Simple(bufPtr *[]byte, num int) {
	buf := *bufPtr
	for i := 0; i < num; i++ {
		buf[i] = ^buf[i]
	}
}
