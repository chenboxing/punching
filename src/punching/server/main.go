package server

import (
	"fmt"
	"log"
	"punching/util"
	"time"
)


var (
	Dch chan bool
	Rch chan util.PairPackage
	Wch chan []byte
)


func Main() {

	// 加载配置信息
	if err := InitConfig(); err != nil {
		fmt.Println("加载配置信息出错，原因为:%s", err)
		return
	}

	// Proxy Server Address
	proxyAddr := Config.Dial
	if proxyAddr == "" {
		proxyAddr = ThirdConfig.Address
	}
	pairName := Config.Key

	var connPeer  util.NetConn


	// 如果跟Peer连接出错，要重新连接Proxy
	for {

		conn, err := ServerDialProxy(proxyAddr, pairName)

		if err != nil {
			log.Println(err)
			time.Sleep(5 * time.Second)
			continue
		}

		localAddr, remoteAddr, _, errWait := WaitForPeer(conn)
		if errWait != nil {
			log.Println(errWait)
			time.Sleep(5 * time.Second)
			continue
		}

		//连接对方
		var errPeer error
		connPeer, errPeer = util.DialPeer(localAddr, remoteAddr)
		if errPeer != nil { //无法连接上
			log.Println(errPeer)
			continue
		}
		
		//已经连接上

		// 连接要开启的服务
		Dch = make(chan bool)
		Rch = make(chan util.PairPackage)
		Wch = make(chan []byte)

		go RHandler(connPeer)   //Nat端写通道
		go WHandler(connPeer)   //Nat端读通道

		// 如果P2P端通讯出错，退出
		select {
		case <-Dch:
			continue
		}


	}



}


func RHandler(conn util.NetConn) {

	//声明一个临时缓冲区，用来存储被截断的数据
	tmpBuffer := make([]byte, 0)

	buff := make([]byte, 1024)
	for {
		j, err := conn.Read(buff)
		if err != nil {
			log.Println("读取连接数据出错，原因为:", err.Error())
			Dch <- true
			break
		}
		log.Println("准备解包数据:", j)
		// 解包
		tmpBuffer = util.UnpackageNat(append(tmpBuffer, buff[:j]...), Rch)
	}
}

func WHandler(conn util.NetConn) {
	for {
		select {
		case msg := <-Wch:
			l, err := conn.Write(msg)
			if err != nil {
				log.Println("写到Nat目录连接出错:", err.Error())
				Dch <- true
			} else {
				log.Println(time.Now().UnixNano(), "已写入到Nat：", l)
			}
			// }

		}
	}
}
