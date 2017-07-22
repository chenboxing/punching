package client

import (
	"fmt"
	"log"
)

func Main() {

	// 加载配置信息
	if err := InitConfig(); err != nil {
		fmt.Println("加载配置信息出错，原因为:%s", err)
		return
	}

	proxyAddr := Config.Dial || ThirdConfig.Address
	pairName := Config.Key
	localAddr, destAddr, pairname, err := util.ClientDialProxy(proxyAddr, pairName)

	if err != nil {
		log.Println(err)
		return
	}

	//连接对方
	connPeer, errPeer := util.DialPeer(localAddr, destAddr)
	if errPeer != nil { //无法连接上
		log.Println(errPeer)
		return
	}

	// 连接上 P2P客户端

	Dch = make(chan bool)
	Rch = make(chan DataPackage)
	Wch = make(chan []byte)

	go RHandler(connPeer)
	go WHandler(connPeer)

	ClientListenHandle()
}

func RHandler(conn util.NetConn) {

}

func WHandler(conn util.NetConn) {

}
