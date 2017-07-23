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
	ProxyDch chan util.ProxyPackage
	ProxyRch chan []byte
	ProxyWch chan []byte
)



// 等待，直到P2P客户端连入，在此期间会一直接受Proxy解析端发来的心跳包
func WaitForPeer(conn util.NetConn) (localAddr string, remoteAddr string, pairName string, err error) {

	defer conn.Close()

	// 接收心跳和客户端连接确认
	ProxyRch = make(chan []byte)
	ProxyWch = make(chan []byte)

	go RProxyHandler(conn)

	select {
	case pack := <-ProxyDch:
		switch pack.CotnrolID {
		case PROXY_CONTROL_QUIT: //无法跟proxy连接
			//关闭连接
			err = error("出现错误")
			break
		case PROXY_CONTROL_NORMAL: // 获取客户端发来信息
			data := pack.Data
			str := string(data)
			parts := strings.Split(str,",")
			localAddr =  parts[0]
			remoteAddr = parts[1]
			pairName = parts[2]
			break
		}
		return
	}
}

// ServerDialProxy P2P服务端连接Proxy解析
func ServerDialProxy(proxyAddr string, pairName string) (retConn util.NetConn, err error) {

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
		data := make([]byte, 512)
		i, _ := conn.Read(data)
		if i == 0 {
			ProxyDch <- util.ProxyPackage{CotnrolID:PROXY_CONTROL_QUIT}
			return
		}

		// Invalid package
		pack := util.UnpackageProxy(data[0:i])
		if pack.Head !=  PROXY_PACKAGE_HEAD {
			ProxyDch <- util.ProxyPackage{CotnrolID:PROXY_CONTROL_QUIT}
			return
		}

		if pack.CotnrolID == PROXY_CONTROL_HEARTBIT {
			//  Received heartbeat package
			// 确认
			ackPack := util.PackageProxy(PROXY_CONTROL_HEARTBITACK, []byte(""))
			conn.Write(ackPack)

		}

		ProxyDch <- pack

	}

}
