package server

import (
	"fmt"
	"log"
	"punching/util"
	"time"
)

func Main() {

	// 加载配置信息
	if err := InitConfig(); err != nil {
		fmt.Println("加载配置信息出错，原因为:%s", err)
		return
	}

	proxyAddr := Config.Dial || ThirdConfig.Address
	pairName := Config.Key

	// 如果跟Peer连接出错，要重新连接Proxy
	for {

		conn, err := ServerDialProxy(proxyAddr, pairName)

		if err != nil {
			log.Println(err)
			time.Sleep(5 * time.Second)
			continue
		}

		localAddr, remoteAddr, errWait := WaitForPeer(conn)
		if errWait != nil {
			log.Println(errWait)
			time.Sleep(5 * time.Second)
			continue
		}

		//连接对方
		connPeer, errPeer := util.DialPeer(localAddr, destAddr)
		if errPeer != nil { //无法连接上
			log.Println(errPeer)
			continue
		}
		
		//已经连接上

	}

}
