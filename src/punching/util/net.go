package util

import (
	"fmt"
	"log"
	"net"
	"time"
	. "punching/constant"
	"strings"
)

// 连接到Proxy代理解析端， 地址/类型/IP
// 发送第一个包->(接收到错误数据，退出，否则接收到NatAddress,->确认收到 <->完成 不断接收心跳包，确认)
// 客户端   错误失败 | nat地址
// 服务端   错误失败 | 注册   心跑包

// ClientDialProxy P2P客户端连接到Proxy
// 连接成功后获取本地地址，远程地址和匹配码，否则将返回错误
func ClientDialProxy(proxyAddr string, pairName string) (localAddr string, remoteAddr string, rePairName string, err error) {

	var conn = NetConn{}

	// 不指定端口，让系统自动分配
	err = conn.Bind("tcp", "")
	if err != nil {
		fmt.Println("绑定出错", err.Error())
		return
	}

	// 连接到Proxy解析服务器
	tcpAddr, err := net.ResolveTCPAddr("tcp", proxyAddr)
	err = conn.Connect(InetAddr(tcpAddr.IP.String()), tcpAddr.Port)

	if err != nil {
		fmt.Println("连接服务端出错", err.Error())
		return
	}
	defer conn.Close()
	fmt.Println("已连接服务器，服务器地址是：%s:%d", tcpAddr.IP.String(), tcpAddr.Port)

	// 构造自定义包
	data := make([]byte, 4)
	data = append(data, []byte{ROLE_CLIENT}...)
	data = append(data, []byte(pairName)...)

	packFirst := PackageProxy(PROXY_CONTROL_FIRST, data)

	_, err = conn.Write(packFirst)
	if err != nil {
		return
	}

	buff := make([]byte, 1024)

	// 获取返回信息
	i, err := conn.Read(buff)
	if err != nil {
		fmt.Println("读取数据出错，", err.Error())
		return
	}

	controlID := buff[1]
	switch controlID {
	case PROXY_CONTROL_NORMAL:
		retData := string(buff[2:i])
		items := strings.Split(retData, ",")
		localAddr = items[0]
		remoteAddr = items[1]
		rePairName = items[2]

		// 发送确认
		packAck := PackageProxy(PROXY_CONTROL_ACK, []byte(""))
		conn.Write(packAck)
		break
	case PROXY_CONTROL_ERROR_NO_SERVER:
		err = fmt.Errorf("错误，P2P服务端不存在")
		break
	case PROXY_CONTROL_ERROR_CLIENT_EXIST:
		err = fmt.Errorf("错误，P2P服务端不存在")
		break
	default:
		err = fmt.Errorf("无效的控制码,%d",int(controlID))
	}

	return
}

func DialPeer(localAddr string, remoteAddr string) (netconn NetConn, err error) {

	remoteTCPAddr, err := net.ResolveTCPAddr("tcp", remoteAddr)
	if err != nil {
		log.Println("The format of remote address is invalid, %s", err.Error())
		return
	}

	var conn NetConn

	// 不指定端口，让系统自动分配
	err = conn.Bind("tcp", localAddr)
	if err != nil {
		log.Println("绑定出错", err.Error())
		return
	}

	log.Println("远程地址是:", remoteTCPAddr.IP.String(), remoteTCPAddr.Port)

	// 有时连接一次并不成功，尝试多次连接
	tryCount := 0
	for {
		tryCount += 1

		if tryCount > 10 {
			err = fmt.Errorf("Attempt to connect remote address, but failed, local addrss: %s, "+
				"remote address", localAddr, remoteAddr)
			return
		}
		err02 := conn.Connect(InetAddr(remoteTCPAddr.IP.String()), remoteTCPAddr.Port)
		if err02 != nil {
			log.Printf("第%d次不能连接远程服务器:%s", tryCount, err02.Error())
			time.Sleep(1 * time.Second)
			continue
		} else {
			log.Println("已经连接到peer: ", remoteTCPAddr.String())
			break
		}
	}
	return conn, nil
}
