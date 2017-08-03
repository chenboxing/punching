package server

import (
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
		logger.Errorf("加载配置信息出错，原因为:%s\n", err)
		return
	}

	// Proxy Server Address
	proxyAddr := Config.Proxy
	if proxyAddr == "" {
		proxyAddr = ThirdConfig.Address
	}
	pairName := Config.Key

	var connPeer util.NetConn

	// 如果跟Peer连接出错，要重新连接Proxy
	for {

		logger.Infof("准备连接Proxy:%s", proxyAddr)
		conn, err := ServerDialProxy(proxyAddr, pairName)

		if err != nil {
			logger.Errorf("连接到Proxy出现错误,", err)
			time.Sleep(5 * time.Second)
			continue
		}

		localAddr, remoteAddr, errWait := WaitForPeer(conn)

		if errWait != nil {
			logger.Errorf("服务端在等待P2P客户端连入出错，原因为:", errWait)
			time.Sleep(5 * time.Second)
			continue
		}

		logger.Infof("服务端：本地地址:%s,对方地址：%s,准备连接", localAddr, remoteAddr)
		//连接对方
		var errPeer error
		connPeer, errPeer = util.DialPeer(localAddr, remoteAddr)
		if errPeer != nil { //无法连接上
			logger.Errorf("无法连接对方,本地地址:%s,远程地址:%s,错误:%s", localAddr, remoteAddr, errPeer)
			continue
		}

		//已经连接上

		// 连接要开启的服务
		Dch = make(chan bool)
		Rch = make(chan util.PairPackage)
		Wch = make(chan []byte)

		go RHandler(connPeer) //Nat端写通道
		go WHandler(connPeer) //Nat端读通道

		// 转发到提供服务端口，并将服务端口数据转到Nat端
		handleServerConn()

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
			logger.Errorf("读取连接数据出错，原因为:", err)
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
				logger.Errorf("写到Nat目录连接出错:", err)
				Dch <- true
			} else {
				logger.Info(time.Now().UnixNano(), "已写入到Nat：", l)
			}

		}
	}
}
