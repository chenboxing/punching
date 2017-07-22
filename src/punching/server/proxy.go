package server

import (
	"fmt"
	"log"
	"net"
	"os"
)

func WaitForPeer(conn util.NetConn) (localAddr string, remoteAddr string, pairname string, err error) {

	// 接收心跳和客户端连接确认
	Rch = make(chan DataPackage)
	Wch = make(chan []byte)

	go RHandler(conn)
	go WHandler(conn)

	select {
	case ret := <-Dch:
		switch ret {
		case 10: //无法跟proxy连接
			//关闭连接
			err = error("出现错误")
			break
		case 11: // 获取客户端发来信息
			localAddr = ""
			remoteAddr = ""
			pairName = ""
			break

		}
		return
	}
}

func ServerDialProxy(proxyAddr string, pairkey string) (conn util.NetConn, err error) {

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

	// 发送第一个包
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

	// 读取
	// 错误，退出此次连接
	// 没有错误，显示当前连接的信息
	i, err := conn.Read(buff01)
	if err != nil {
		fmt.Println("读取数据出错，", err.Error())
		os.Exit(1)
	}

	// log.Print("读取数据:", string(buff01[0:i]))
	// if string(buff01[0:i]) == key {
	// 	break
	// }
	// log.Println("发送数据", []byte(name))
	// conn.Write([]byte(name))

}

func RHandler(conn util.NetConn) {

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
