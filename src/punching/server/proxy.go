package server

import (
	"fmt"
	"log"
	"net"
	"punching/util"
	. "punching/constant"
	"strings"
)
var (
	ProxyDch chan byte
	ProxyRch chan []byte
	ProxyWch chan []byte
)


func WaitForPeer(conn util.NetConn) (localAddr string, remoteAddr string,  err error) {

	// 接收心跳和客户端连接确认
	ProxyRch = make(chan []byte)
	ProxyWch = make(chan []byte)

	go RProxyHandler(conn)


	select {
	case ret := <-ProxyDch:
		switch ret {
		case PROXY_CONTROL_QUIT: //无法跟proxy连接
			//关闭连接
			err = error("出现错误")
			break
		case PROXY_CONTROL_NORMAL: // 获取客户端发来信息
			localAddr = ""
			remoteAddr = ""
			pairName = ""
			break

		}
		return
	}
}

func ServerDialProxy(proxyAddr string, pairName string) (retConn util.NetConn, err error) {

	// 发第一个包
	// 地址接收确认
	// 心跑包接收和确认

	var conn util.NetConn

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

	fmt.Println("已连接服务器，服务器地址是：%s:%d", tcpAddr.IP.String(), tcpAddr.Port)


	// 构造自定义包
	data := make([]byte, 4)
	data = append(data, []byte{ROLE_SERVER}...)
	data = append(data, []byte(pairName)...)

	packFirst := util.PackageProxy(PROXY_CONTROL_FIRST, data)

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
		localAddr := items[0]
		rePairName := items[1]

		fmt.Printf("P2P服务端侦听地址为：%s, 匹配码为:%s", localAddr, rePairName)

		break
	case PROXY_CONTROL_ERROR_SERVER_EXIST:
		err = fmt.Errorf("错误，P2P服务端已存在")
		break
	default:
		err = fmt.Errorf("无效的控制码,%d",int(controlID))
	}

	return conn, nil

}

func RProxyHandler(conn util.NetConn) {

	for {
		// 心跳包,回复ack
		data := make([]byte, 128)
		i, _ := conn.Read(data)
		if i == 0 {
			Dch <- true
			return
		}
		if data[0] == Req_HEARTBEAT {
			fmt.Println("recv ht pack")
			conn.Write([]byte{Res_REGISTER, '#', 'h'})
			fmt.Println("send ht pack ack")
		} else if data[0] == Req { // 接收到确认信息
			fmt.Println("recv data pack")
			fmt.Printf("%v\n", string(data[2:]))
			conn.Write([]byte{Res, '#'})
		}
	}

}
