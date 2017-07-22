package client

import (
	"fmt"
	"log"
	"punching/util"
	"time"
	"os"
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

	proxyAddr := Config.Dial
	if proxyAddr == "" {
		proxyAddr = ThirdConfig.Address
	}
	pairName := Config.Key
	localAddr, destAddr, pairName, err := util.ClientDialProxy(proxyAddr, pairName)

	if err != nil {
		log.Println(err)
		return
	}

	//连接P2P服务端
	connPeer, errPeer := util.DialPeer(localAddr, destAddr)
	if errPeer != nil { //无法连接上
		log.Println(errPeer)
		return
	}

	Dch = make(chan bool)
	Rch = make(chan util.PairPackage)
	Wch = make(chan []byte)

	go RHandler(connPeer)   //Nat端写通道
	go WHandler(connPeer)   //Nat端读通道

	// 侦听端口，开启服务，将端口输入转发到P2P端
	ClientListenHandle()

	// 如果P2P端通讯出错，退出
	select {
	case <-Dch:
		os.Exit(1)
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
			// prefix := string(msg[0:2])
			// if prefix != "00" && prefix != "01" && prefix != "11" {
			// 	log.Println("not equal", string(msg))
			// } else {
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
