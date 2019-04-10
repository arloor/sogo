# sogo

一个使用http流量进行混淆的socks5代理。

服务器端：[sogo-server](https://github.com/arloor/sogo-server)

之前写了一个http代理，用起来也是十分地舒服，但是有几个点还是有些遗憾的：

- http代理只能代理http协议，相比socks5代理不够通用。。
- netty是个好框架，但是java占用内存是真的多。。

所以，我又写了一个socks5代理，起名叫[sogo](https://github.com/arloor/sogo)。

sogo本身包含sogo(client)和sogo-server。如果把sogo和sogo-server看成一个整体，一个黑盒，这个整体就是一个socks5代理。sogo(client)与本地电脑交互；sogo-server与目标网站交互；sogo(client)和sogo-server之间的交互就是http协议包裹payload进行通信。sogo(client)和sogo-server之间的这段就是翻墙的重点——采取各种方式混过GFW，我采用的“http流量包裹payload”应该算是比较优雅的一种。

## 运行日志

以下是观看一个youtube视频的日志：

```shell
2019/04/10 00:14:28 main.go:58: 与客户端握手成功
2019/04/10 00:14:28 main.go:207: 目的地址类型:3 域名长度:15 目标域名:www.youtube.com 目标端口:443
2019/04/10 00:14:28 main.go:58: 与客户端握手成功
2019/04/10 00:14:28 main.go:207: 目的地址类型:3 域名长度:11 目标域名:i.ytimg.com 目标端口:443
2019/04/10 00:14:28 main.go:58: 与客户端握手成功
2019/04/10 00:14:28 main.go:207: 目的地址类型:3 域名长度:13 目标域名:yt3.ggpht.com 目标端口:443
2019/04/10 00:14:35 main.go:58: 与客户端握手成功
2019/04/10 00:14:35 main.go:207: 目的地址类型:3 域名长度:32 目标域名:r2---sn-i3belnel.googlevideo.com 目标端口:443
2019/04/10 00:14:35 main.go:58: 与客户端握手成功
2019/04/10 00:14:35 main.go:207: 目的地址类型:3 域名长度:32 目标域名:r2---sn-i3belnel.googlevideo.com 目标端口:443
2019/04/10 00:14:35 main.go:58: 与客户端握手成功
2019/04/10 00:14:35 main.go:207: 目的地址类型:3 域名长度:32 目标域名:r2---sn-i3belnel.googlevideo.com 目标端口:443
```

以下是访问github的日志：
```shell
2019/04/10 00:15:57 main.go:58: 与客户端握手成功
2019/04/10 00:15:57 main.go:207: 目的地址类型:3 域名长度:10 目标域名:github.com 目标端口:443
2019/04/10 00:15:57 main.go:58: 与客户端握手成功
2019/04/10 00:15:57 main.go:207: 目的地址类型:3 域名长度:10 目标域名:github.com 目标端口:443
2019/04/10 00:15:59 main.go:58: 与客户端握手成功
2019/04/10 00:15:59 main.go:207: 目的地址类型:3 域名长度:15 目标域名:live.github.com 目标端口:443
2019/04/10 00:16:00 main.go:58: 与客户端握手成功
2019/04/10 00:16:00 main.go:207: 目的地址类型:3 域名长度:14 目标域名:api.github.com 目标端口:443
```



## 特性

sogo项目最好的两个特性如下：

1. 使用http包裹payload(有意义的数据)。
2. 将sogo-server所在的ip:端口伪装成一个http网站。

效用、坚固、美观——对软件产品的三个要求。上面两个特性，既可以说是坚固，也可以说是美观，至于效用就不用说了，在这里谈坚固和美观的前提就是效用被完整地实现。用通俗地话来说，这个代理的坚固和美观就是：伪装、防止被识别。

## 处理socks5握手——对socks5协议的实现

sogo(client)与本地电脑交互，因此需要实现socks5协议，与本地用户（比如chrome）握手协商。

一个典型的sock5握手的顺序：

1. client：0x05 0x01 0x00
2. server: 0x05 0x00 
3. client: 0x05 0x01 0x00 0x01 ip1 ip2 ip3 ip4 0x00 0x50
4. server: 0x05 0x00 0x00 0x01 0x00 0x00 0x00 0x00 0x10 0x10
5. client与server开始盲转发

这一部分代码见如下两个函数：

```java
//file: sogo/main.go

//读 5 1 0 写回 5 0
func handshake(clientCon net.Conn) error {
	var buf = make([]byte, 300)
	numRead, err := clientCon.Read(buf)
	if err != nil {
		return err
	} else if numRead == 3 && buf[0] == 0X05 && buf[1] == 0X01 && buf[2] == 0X00 {
		return mio.WriteAll(clientCon, []byte{0x05, 0x00})
	} else {
		log.Printf("%d", buf[:numRead])
		return mio.WriteAll(clientCon, []byte{0x05, 0x00})
	}
}

func getTargetAddr(clientCon net.Conn) (string, error) {
	var buf = make([]byte, 1024)
	numRead, err := clientCon.Read(buf)
	if err != nil {
		return "", err
	} else if numRead > 3 && buf[0] == 0X05 && buf[1] == 0X01 && buf[2] == 0X00 {
		if buf[3] == 3 {
			log.Printf("目的地址类型:%d 域名长度:%d 目标域名:%s 目标端口:%s", buf[3], buf[4], buf[5:5+buf[4]], strconv.Itoa(int(binary.BigEndian.Uint16(buf[5+buf[4]:7+buf[4]]))))
			writeErr := mio.WriteAll(clientCon, []byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x10, 0x10})
			return string(buf[5:5+buf[4]]) + ":" + strconv.Itoa(int(binary.BigEndian.Uint16(buf[5+buf[4]:7+buf[4]]))), writeErr
		} else if buf[3] == 1 {
			log.Printf("目的地址类型:%d  目标域名:%s 目标端口:%s", buf[3], net.IPv4(buf[4], buf[5], buf[6], buf[7]).String(), strconv.Itoa(int(binary.BigEndian.Uint16(buf[8:10]))))
			writeErr := mio.WriteAll(clientCon, []byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x10, 0x10})
			return net.IPv4(buf[4], buf[5], buf[6], buf[7]).String() + ":" + strconv.Itoa(int(binary.BigEndian.Uint16(buf[8:10]))), writeErr
		} else {
			return "", errors.New("不能处理ipv6")
		}

	} else {
		return "", errors.New("不能处理非CONNECT请求")
	}
}
```

完成handshake, getTargetAddr 之后，chrome就会发送真实的http请求了，sogo(client) 要做的就是将这部分http请求进行加密，然后加上http请求的头，发送到sogo-server。

## 使用http包裹payload

第一部分：如何将真实的http请求，再进行加密，最后加上假的http请求头，变成伪装好的http请求，发送给sogo-server。


```java
//file: sogo/mio/prefix.go
var fakeHost = "qtgwuehaoisdhuaishdaisuhdasiuhlassjd.com"  //虚假host

func AppendHttpRequestPrefix(buf []byte, addr string) []byte {
	Simple(&buf, len(buf))//对真实的http请求的简单加密
	// 演示base64编码
	addrBase64 := base64.NewEncoding("abcdefghijpqrzABCKLMNOkDEFGHIJl345678mnoPQRSTUVstuvwxyWXYZ0129+/").EncodeToString([]byte(addr))
	buf = append([]byte("POST /target?at="+addrBase64+" HTTP/1.1\r\nHost: "+fakeHost+"\r\nAccept: */*\r\nContent-Type: text/plain\r\naccept-encoding: gzip, deflate\r\ncontent-length: "+strconv.Itoa(len(buf))+"\r\n\r\n"), buf...)
	return buf
}
```

包裹完毕之后返回的[]byte就可以发送给sogo-server了。

第二部分：将sogo-server从目标网站获得的真实响应进行简单加密，包裹http响应头，发送给sogo(client)。




```java
//file: sogo-server/mio/prefix.go
func AppendHttpResponsePrefix(buf []byte) []byte {
	Simple(&buf, len(buf))
	buf = append([]byte("HTTP/1.1 200 OK\r\nContent-Type: text/plain; charset=utf-8\r\nContent-Length: "+strconv.Itoa(len(buf))+"\r\n\r\n"), buf...)
	return buf
}
```

包裹完毕之后返回的[]byte就可以发送给sogo(client)了。

## 解包伪装好的请求、响应

先看以下伪装好的请求的样子：

```shell
POST /target?at={targetAddrBase64} HTTP/1.1
Host: {fakehost}
Accept: */*
Content-Type: text/plain
accept-encoding: gzip, deflate
content-length: {content-length}

{payload-after-crypto}
```

sogo-server拿到这个伪装好的请求，要做的事有：

1. 获取{targetAddrBase64}，拿到真实的目标网站地址
2. 获取请求头的Host字段，如果不是定义好的fakehost，则说明是直接访问sogo-server，这时sogo-server就是个到混淆网站的反向代理（这就是之前提到的第二个特性。下面将会详细解释如何实现
3. 获取{content-length}，根据这个content-length确定payload部分的长度。
4. 读取指定长度的payload，解密，并创建到targetAddr的连接，转发至targetAddr

这些步骤很明确吧。其实有一些细节，挺麻烦的。

tcp是面向流的协议，也就是会有很多个连续的上面的片段，要合理划分出这些片段。有些人称这个为解决“tcp粘包”，谷歌tcp粘包就能搜到如何实现这个需求。但是注意，不要称这个为“tcp粘包”，别人会说tcp是面向流的协议，哪来什么包，你知识体系有问题，你看过tcp协议没有。这些话都是知乎上某一问题的答案说的。所以，别说“tcp粘包”，但是可以用这个关键词去搜索如何解决这个问题。

如果，现在你看了如何解决这个问题，其实就是一句话，在tcp上层定义自己的应用层协议：也就是tcp报文的格式。http这个应用层协议就是一种tcp报文的一种定义。

我们的伪装好的报文就是http协议，所以要做的就是实现自己的http请求解析器，获取我们关心的信息。

sogo的http请求解析器，在：

```java
//file sogo-server/server.go
func read(clientConn net.Conn, redundancy []byte) (payload, redundancyRetain []byte, target string, readErr error)
```

这一部分有点繁杂。。不多解释，自己看代码吧。

## 伪装sogo-server:80为其他http网站

这一部分就是第二特性：将sogo-server所在的ip:端口伪装成一个http网站。

上一节，我们提到 {fakehost}。我们故意将{fakehost}定义为一个复杂、很长的域名。我们伪装的请求，都会带有如下请求头

```shell
Host: {fakehost}
```

如果，http请求的Host不是这个{fakehost}则说明这不是一个经sogo(client)的请求，而是直接请求了sogo-server。也就是，有人来嗅探啦！

对这种，我们就会将该请求，原封不动地转到伪装站。（其实还是有点修改的，但这是细节，看代码吧）所以，直接访问sogo-server-ip:80 就是访问伪装站：80。


## 结束

这篇博客梳理了一下sogo的实现原理，总之，sogo是一个优雅的翻墙代理。并且机缘巧合也获得了一些实际的好处，还是挺舒服的。sogo代码不多，对go语言、翻墙原理、网络编程感兴趣的人可以看看。

