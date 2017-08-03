package client

import (
	"os"
	"punching/logger"
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
		logger.Errorf("加载配置信息出错，原因为:%s", err)
		return
	}

	proxyAddr := Config.Proxy
	if proxyAddr == "" {
		proxyAddr = ThirdConfig.Address
	}
	pairName := Config.Key

	logger.Infof("准备连接代理解析端:%s", proxyAddr)

	tryCount := 0
	var connPeer util.NetConn
	var errPeer error

	for {

		localAddr, destAddr, err := util.ClientDialProxy(proxyAddr, pairName)

		if err != nil {
			logger.Errorf("连接解析端出错,%s", err)
			return
		}

		logger.Infof("已获取NAT地址：本地地址:%s，远程地址:%s ", localAddr, destAddr)

		tryCount += 1

		if tryCount == 11 {
			logger.Errorf("已经尝试了10次，连接还是失败，退出，请重新运行客户端")
			return
		}
		//连接P2P服务端
		connPeer, errPeer = util.DialPeer(localAddr, destAddr)
		if errPeer != nil { //无法连接上
			logger.Errorf("连接P2P服务端，出现错误,%s,第%d次", errPeer, tryCount)
			time.Sleep(3 * time.Second)
			continue
		} else {
			break
		}

	}

	Dch = make(chan bool)
	Rch = make(chan util.PairPackage)
	Wch = make(chan []byte)

	go RHandler(connPeer) //Nat端写通道
	go WHandler(connPeer) //Nat端读通道

	// 侦听端口，开启服务，将端口输入转发到P2P端
	ClientListenHandle()

	// 如果P2P端通讯出错，退出
	select {
	case <-Dch:
		logger.Errorf("接收到退出信息")
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
			logger.Errorf("读取连接数据出错，原因为:%s", err)
			Dch <- true
			break
		}
		logger.Info("准备解包数据:", j)
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
				logger.Errorf("写到Nat目录连接出错:%s", err)
				Dch <- true
			} else {
				logger.Info(time.Now().UnixNano(), "已写入到Nat：", l)
			}
			// }

		}
	}
}
