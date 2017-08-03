package server

import (
	"fmt"

	"errors"
	"net"
	. "punching/constant"
	"punching/logger"
	"punching/util"
	"strings"
)

var (
	ProxyDch chan util.ProxyPackage
	ProxyRch chan []byte
	ProxyWch chan []byte
)

// 等待，直到P2P客户端连入，在此期间会一直接受Proxy解析端发来的心跳包
func WaitForPeer(conn *util.NetConn) (localAddr string, remoteAddr string, err error) {
	defer conn.Close()
	logger.Infof("Enter WaitForPeer")

	// 接收心跳和客户端连接确认
	ProxyRch = make(chan []byte)
	ProxyWch = make(chan []byte)

	ProxyDch = make(chan util.ProxyPackage)

	go RProxyHandler(conn)

	select {
	case pack := <-ProxyDch:

		switch pack.ControlID {
		case PROXY_CONTROL_QUIT: //无法跟proxy连接
			//关闭连接
			logger.Error("收到退出包")
			err = fmt.Errorf("收到退出")
			return
		case PROXY_CONTROL_NORMAL: // 获取客户端发来信息
			logger.Info("读取到客户端发来的信息包")
			data := pack.Data
			str := string(data)
			parts := strings.Split(str, ",")
			remoteAddr = parts[0]
			localAddr  = parts[1]
			return
		}
	}

	return
}

// ServerDialProxy P2P服务端连接Proxy解析
func ServerDialProxy(proxyAddr string, pairName string) (retConn *util.NetConn, err error) {

	var conn = &util.NetConn{}

	// 不指定端口，让系统自动分配
	err = conn.Bind("")
	if err != nil {
		logger.Errorf("绑定出错%s", err)
		return
	}

	// 连接到Proxy解析服务器
	tcpAddr, err := net.ResolveTCPAddr("tcp", proxyAddr)

	err = conn.Connect(util.InetAddr(tcpAddr.IP.String()), tcpAddr.Port)

	if err != nil {
		logger.Errorf("连接服务出错:%s", proxyAddr)
		fmt.Println("连接服务端出错", err.Error())
		return
	}

	logger.Infof("已连接服务器，服务器地址是：%s:%d", tcpAddr.IP.String(), tcpAddr.Port)

	// 构造自定义包
	data := make([]byte, 0)
	data = append(data, []byte{ROLE_SERVER}...)
	data = append(data, []byte(pairName)...)

	packFirst := util.PackageProxy(PROXY_CONTROL_FIRST, data)

	_, err = conn.Write(packFirst)

	if err != nil {
		logger.Errorf("写入Proxy连接出错,%s", err)
		return
	}

	buff := make([]byte, 1024)

	// 获取返回信息
	i, err := conn.Read(buff)
	if err != nil {
		logger.Errorf("读取数据出错，%s", err)
		return
	}

	controlID := buff[1]
	switch controlID {
	case PROXY_CONTROL_NORMAL:
		retData := string(buff[2:i])
		items := strings.Split(retData, ",")
		localAddr := items[0]
		rePairName := items[1]

		logger.Infof("P2P服务端侦听地址为：%s, 匹配码为:%s", localAddr, rePairName)
		break
	case PROXY_CONTROL_ERROR_SERVER_EXIST:
		logger.Error("错误，P2P服务端已存在")
		err = errors.New("错误，P2P服务端已存在")
		break
	default:
		err = fmt.Errorf("无效的控制码,%d", int(controlID))
	}

	return conn, err

}

func RProxyHandler(conn *util.NetConn) {

	for {
		// 心跳包,回复ack
		data := make([]byte, 512)
		i, err0 := conn.Read(data)
		if err0 != nil {
			logger.Errorf("读取Proxy连接出错，%s", err0)
			ProxyDch <- util.ProxyPackage{ControlID: PROXY_CONTROL_QUIT}
			return
		}

		// Invalid package
		pack, err := util.UnpackageProxy(data[0:i])
		if err != nil {
			logger.Errorf("解包错误,%s", err)
			ProxyDch <- util.ProxyPackage{ControlID: PROXY_CONTROL_QUIT}
			return
		}

		if pack.ControlID == PROXY_CONTROL_HEARTBIT {
			//  Received heartbeat package
			// 确认
			ackPack := util.PackageProxy(PROXY_CONTROL_HEARTBITACK, []byte(""))
			conn.Write(ackPack)

		} else {

			ProxyDch <- pack
			return

		}

	}

}
