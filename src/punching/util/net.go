package util

import (
	"fmt"
	"net"
	. "punching/constant"
	"punching/logger"
	"strings"
)

// ClientDialProxy P2P客户端连接到Proxy端
// 连接成功后获取本地地址，远程地址和匹配码，否则将返回错误
func ClientDialProxy(proxyAddr string, pairName string) (localAddr string, remoteAddr string, err error) {

	var conn = NetConn{}

	// 不指定端口，让系统自动分配
	err = conn.Bind("")
	if err != nil {
		logger.Errorf("绑定出错,%s", err)
		return
	}

	// 连接到Proxy解析服务器
	tcpAddr, err := net.ResolveTCPAddr("tcp", proxyAddr)
	err = conn.Connect(InetAddr(tcpAddr.IP.String()), tcpAddr.Port)

	if err != nil {
		logger.Errorf("连接服务端出错,%s", err)
		return
	}
	defer conn.Close()
	logger.Infof("已连接服务器，服务器地址是：%s:%d", tcpAddr.IP.String(), tcpAddr.Port)

	// 构造自定义包
	data := make([]byte, 0)
	data = append(data, []byte{ROLE_CLIENT}...)
	data = append(data, []byte(pairName)...)

	packFirst := PackageProxy(PROXY_CONTROL_FIRST, data)

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
		remoteAddr = items[0]
		localAddr = items[1]

		//// 发送确认
		//packAck := PackageProxy(PROXY_CONTROL_ACK, []byte(""))
		//conn.Write(packAck)
		return
	case PROXY_CONTROL_ERROR_NO_SERVER:
		err = fmt.Errorf("错误，P2P服务端不存在")
		break
	case PROXY_CONTROL_ERROR_CLIENT_EXIST:
		err = fmt.Errorf("错误，P2P客户端已存在")
		break
	default:
		err = fmt.Errorf("无效的控制码,%d", int(controlID))
	}

	return
}

func DialPeer(localAddr string, remoteAddr string) (netconn NetConn, err error) {

	remoteTCPAddr, err := net.ResolveTCPAddr("tcp", remoteAddr)
	if err != nil {
		logger.Errorf("The format of remote address is invalid, %s", err)
		return
	}
	var conn NetConn

	// 不指定端口，让系统自动分配
	err = conn.Bind(localAddr)
	if err != nil {
		logger.Errorf("绑定出错,%s", err)
		return
	}

	// 有时连接一次并不成功，尝试多次连接
	//tryCount := 0
	//for {
	//	tryCount += 1
	//
	//	if tryCount > 10 {
	//		err = fmt.Errorf("Attempt to connect remote address, but failed, local addrss: %s, "+
	//			"remote address:%s", localAddr, remoteAddr)
	//		return
	//	}
	err = conn.Connect(InetAddr(remoteTCPAddr.IP.String()), remoteTCPAddr.Port)
	if err != nil {
		logger.Warnf("第次不能连接远程服务器:%s", err)
		//time.Sleep(1 * time.Second)
		//continue
	} else {
		logger.Infof("已经连接到peer: %s", remoteTCPAddr.String())
		//break
	}
	//}
	return conn, err
}
