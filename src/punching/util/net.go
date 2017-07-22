package util

import (
	"fmt"
	"log"
	"net"
	"os"
	"p2p/util"
	"strings"
	"time"
)

// 连接到Proxy代理解析端， 地址/类型/IP
// 发送第一个包->(接收到错误数据，退出，否则接收到NatAddress,->确认收到 <->完成 不断接收心跳包，确认)
// 客户端   错误失败 | nat地址
// 服务端   错误失败 | 注册   心跑包
func ClientDialProxy(proxyAddr string, pairkey string, role string) (localAddr string, destAddr string, pairname string, err error) {

	// 发第一个包
	// 地址接收确认
	// 心跑包接收和确认

	var conn = util.NetConn{}

	// 不指定端口，让系统自动分配
	err = conn.Bind("tcp", "")
	if err != nil {
		log.Println("绑定出错", err.Error())
		return
	}

	// 连接到Proxy解析服务器
	tcpAddr, err := net.ResolveTCPAddr("tcp", proxyAddr)
	err = conn.Connect(util.InetAddr(tcpAddr.IP.String()), tcpAddr.Port)

	if err != nil {
		fmt.Println("连接服务端出错", err.Error())
		return
	}
	defer conn.Close()
	fmt.Println("已连接服务器，服务器地址是：%s:%d", tcpAddr.IP.String(), tcpAddr.Port)

	for {

		PackageProxy()

		key := "ok" + name
		buff01 := make([]byte, len(key))

		//发送第一个包
		log.Println("发送数据1", name)
		n, err := conn.Write([]byte(name))
		if err != nil {
			log.Println("出错，原因：", err.Error())
			os.Exit(22)
		}
		log.Println("have write: ", n)

		i, err := conn.Read(buff01)
		if err != nil {
			fmt.Println("读取数据出错，", err.Error())
			os.Exit(1)
		}
		log.Print("读取数据:", string(buff01[0:i]))
		if string(buff01[0:i]) == key {
			break
		}
		log.Println("发送数据", []byte(name))
		conn.Write([]byte(name))
	}

	fmt.Println("--- Request sent, waiting for parkner in Name ...", name)

	buff := make([]byte, 512)

	//获取服务发送过来的对方IP信息
	len, err := conn.Read(buff)
	if err != nil {
		log.Println("客户端读取数据出错:", err.Error())
		os.Exit(1)
	}

	pairAddr := string(buff[0:len])
	arrPairAddr := strings.Split(pairAddr, ",")
	remoteAddrStr := arrPairAddr[0]
	localAddrStr := arrPairAddr[1]
	log.Println("读取到远程IP", remoteAddrStr)

	conn.Close()
	//time.Sleep(3 * time.Second)

	remoteAddr, err := net.ResolveTCPAddr("tcp", remoteAddrStr)

	//localAddr.String()
	log.Println("local addr:", localAddrStr)

}



func DialPeer(localAddr string, remoteAddr string) util.NetConn {

	remoteTCPAddr, err := net.ResolveTCPAddr("tcp", remoteAddr)

	var conn util.NetConn

	// 不指定端口，让系统自动分配
	err = conn.Bind("tcp", localAddr)
	if err != nil {
		log.Println("绑定出错", err.Error())
		return
	}

	// TODO 应该封装成一个函数
	//      有时连接并不成功，多次连接
	log.Println("远程地址是:", remoteAddr.IP.String(), remoteAddr.Port)

	tryCount := 0
	for {
		tryCount += 1

		if tryCount > 10 {
			log.Printf("连接不上，退出")
			return
		}
		err02 := conn02.Connect(util.InetAddr(remoteAddr.IP.String()), remoteAddr.Port)
		if err02 != nil {
			log.Printf("第%d次不能连接远程服务器:%s", tryCount, err02.Error())
			time.Sleep(1 * time.Second)
			continue
		} else {
			log.Println("已经连接到peer: ", remoteAddr.String())
			break
		}
	}

}
