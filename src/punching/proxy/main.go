package proxy

import (
	"fmt"

	"net"
	. "punching/constant"
	"punching/logger"
	"punching/util"
	"sync"
	"time"
)

// ServerConn 服务端到代理端连接
type ServerConn struct {
	Rch       chan []byte // 读通道
	Wch       chan []byte // 写通道
	Dch       chan bool   // 连接退
	LocalAddr string      // 客户端IP信息
	Pairname  string      // 匹配名称
	SyncAt    int64       // 上次心跳时间
}

// ClientConn 客户端到代理端连接
type ClientConn struct {
	Pairname string //匹配码
}

// 全局变量
var (
	OnlineServerList map[string]*ServerConn // 服务端连接列表Map
	OnlineClientList map[string]string      // 客户端连接列表Map
	RWLockClient     *sync.RWMutex          //读写锁
	RWLockServer     *sync.RWMutex
)

func Main() {

	// 加载配置信息
	if err := InitConfig(); err != nil {
		logger.Errorf("加载配置信息出错，原因为:%s", err)
		return
	}

	OnlineServerList = make(map[string]*ServerConn)
	OnlineClientList = make(map[string]string)

	RWLockClient = new(sync.RWMutex)
	RWLockServer = new(sync.RWMutex)

	listenAddr := Config.Listen
	tcpAddr, err := net.ResolveTCPAddr("tcp", listenAddr)
	if err != nil {
		logger.Errorf("Resolved address failed %s", err)
		panic(err)
	}

	listen, err := net.ListenTCP("tcp", tcpAddr)

	if err != nil {
		logger.Errorf("监听端口失败:%s", err)
		return
	}
	logger.Infof("已初始化连接，正在侦听:%s, 等待客户端连接...", listenAddr)

	for {
		conn, err := listen.AcceptTCP()
		if err != nil {
			fmt.Println("连接异常:", err.Error())
			continue
		}
		logger.Infof("本地地址:%s,来自远程地址:%s", conn.LocalAddr().String(), conn.RemoteAddr().String())
		go Handler(conn)
	}
}

// processRoleClient 处理客户端连接
func processRoleClient(conn net.Conn, pairName string) {

	// 判断匹配的服务端是否已经注册
	RWLockServer.RLock()

	serverConn, ok := OnlineServerList[pairName]

	RWLockServer.RUnlock()

	if !ok {
		// 客户端没有注册
		packErr := util.PackageProxy(PROXY_CONTROL_ERROR_NO_SERVER, []byte(""))
		conn.Write(packErr)
		return
	}

	// Check the client with the save pair name exists
	RWLockClient.RLock()
	_, ok = OnlineClientList[pairName]
	RWLockClient.RUnlock()

	if ok {
		packErr := util.PackageProxy(PROXY_CONTROL_ERROR_CLIENT_EXIST, []byte(""))
		conn.Write(packErr)
		return
	}

	// 添加到客户端列表
	RWLockClient.Lock()
	OnlineClientList[pairName] = pairName
	RWLockClient.Unlock()

	// 发送Nat地址和接收确认
	toClientAddrs := serverConn.LocalAddr + "," + conn.RemoteAddr().String()

	pack := util.PackageProxy(PROXY_CONTROL_NORMAL, []byte(toClientAddrs))
	conn.Write(pack)

	//buf := make([]byte, 512)
	//lenAck, err := conn.Read(buf)
	//
	//if err != nil {
	//	logger.Errorf("读客户端确认数据出错,%s", err)
	//	return
	//}
	//
	//ackPack, err01 := util.UnpackageProxy(buf[0:lenAck])
	//if err01 != nil {
	//	logger.Errorf("包解析出问题")
	//	return
	//}
	//flag := 0
	//if ackPack.ControlID == PROXY_CONTROL_ACK {
	//	flag += 1
	//}

	toServerAddrs := conn.RemoteAddr().String() + "," + serverConn.LocalAddr
	addrPack := util.PackageProxy(PROXY_CONTROL_NORMAL, []byte(toServerAddrs))

	serverConn.Wch <- addrPack

	//// 等待服务端的确认数据
	//select {
	//case bufAck := <-serverConn.Rch:
	//	pack, err02 := util.UnpackageProxy(bufAck)
	//	if err02 != nil {
	//		if pack.ControlID == PROXY_CONTROL_ACK {
	//			flag += 1
	//		}
	//	}
	//	break
	//}
	//
	//logger.Infof("当前的连接信息为: %d", flag)

	// 收到服务端的确认数据
	//	if flag == 2 {
	RWLockServer.Lock()
	//serverConn.Dch <- true // 关闭服务端连接
	delete(OnlineServerList, pairName)
	RWLockServer.Unlock()
	//	}

	RWLockClient.Lock()
	delete(OnlineClientList, pairName)
	RWLockClient.Unlock()

	return

}

// Handle 连接处理函数
func Handler(conn net.Conn) {

	//defer func() {
	//	if err := recover(); err != nil {
	//		logger.Errorf("连接出现问题:%s", err)
	//	}
	//}()

	defer conn.Close()
	buf := make([]byte, 1024)

	var pairName string

	for {

		i, err := conn.Read(buf)
		if err != nil {
			logger.Errorf("读取数据错误:%s", err)
			return
		}

		firstPack, err01 := util.UnpackageProxy(buf[0:i])

		if err01 != nil {
			logger.Errorf("包格式出错,%s", err01)
			return
		}

		// Todo 获取时间差距  服务器时间 ticks - 客户端时间 ticks
		// 比如说3秒后

		clientType := firstPack.Data[0]

		if len(firstPack.Data) > 1 {
			pairName = string(firstPack.Data[1:])
		}

		// 处理客户端连接
		if clientType == ROLE_CLIENT {
			processRoleClient(conn, pairName)
			return // 退出客户端连接
		}
		break
	}

	// 下面的操作都是针对服务端连接
	logger.Info("处理P2P服务端连接")

	// 服务端连接允许匹配码为空，系统将随机产生唯一匹配码
	if pairName == "" {
		for {
			pairName = util.GenerateRandomPairKey()
			RWLockServer.Lock()
			if _, ok := OnlineServerList[pairName]; !ok {
				break
			}
			RWLockServer.Unlock()
		}
	} else {

		// 是否存在pair name
		RWLockServer.RLock()
		_, ok := OnlineServerList[pairName]
		RWLockServer.RUnlock()

		// 已存在
		if ok {
			errPack := util.PackageProxy(PROXY_CONTROL_ERROR_SERVER_EXIST, []byte(""))
			conn.Write(errPack)
			logger.Errorf("服务端列表中已存在:%s", pairName)
			return
		}

	}

	logger.Infof("匹配码为:%s", pairName)

	// 生成服务器连接对象添加到列表
	RWLockServer.Lock()
	serverConn := &ServerConn{Rch: make(chan []byte),
		Wch:       make(chan []byte),
		Dch:       make(chan bool),
		Pairname:  pairName,
		LocalAddr: conn.RemoteAddr().String()}
	OnlineServerList[pairName] = serverConn
	RWLockServer.Unlock()

	replyData := conn.RemoteAddr().String() + "," + pairName
	replyPack := util.PackageProxy(PROXY_CONTROL_NORMAL, []byte(replyData))
	if _, err := conn.Write(replyPack); err != nil {
		logger.Errorf("回复P2P服务端包出错", err)
		return
	}

	//	写通道
	go WHandler(conn, serverConn)

	//	读通道
	go RHandler(conn, serverConn)

	// 等待退出通道
	select {
	case <-serverConn.Dch:
		logger.Info("close handler goroutine")
	}
}

// 正常写数据 匹配端连接上来会写信息
// 定时检测 conn die => goroutine die
func WHandler(conn net.Conn, C *ServerConn) {

	for {
		select {
		case d := <-C.Wch:
			logger.Info("通道接收到数据，准备写")
			conn.Write(d)
		}
	}
}

// 读客户端数据 + 心跳检测
func RHandler(conn net.Conn, C *ServerConn) {

	// 心跳ack
	// 业务数据 写入Wch

	for {
		data := make([]byte, 128)
		// 设置读超时
		err := conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		if err != nil {
			logger.Errorf("设置读超时失败,%s", err)
		}
		if i, derr := conn.Read(data); derr == nil {
			// 可能是来自客户端的消息确认
			//           	     数据消息
			pack, err01 := util.UnpackageProxy(data[0:i])
			if err01 != nil {
				logger.Errorf("包无法解析,%s", err01)
				continue
			}
			if pack.ControlID == PROXY_CONTROL_HEARTBITACK {
				logger.Info("Received hartbeat ack") //// C.Rch <- data
			}

			continue
		}

		//如果等待10秒没有读到客户端数据或读写出错，写心跳包
		// 写心跳包
		heartPack := util.PackageProxy(PROXY_CONTROL_HEARTBIT, []byte(""))
		conn.Write(heartPack)
		conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		if _, herr := conn.Read(data); herr == nil {

			// fmt.Println(string(data))
			// 更新心跳时间
			RWLockServer.RLock()
			serverConn, ok := OnlineServerList[C.Pairname]
			if ok {
				serverConn.SyncAt = time.Now().Unix()
			}
			RWLockServer.RUnlock()

		} else {
			logger.Errorf("读取连接出错，%s", herr)
			RWLockServer.Lock()
			delete(OnlineServerList, C.Pairname)
			RWLockServer.Unlock()
			logger.Infof("删除在线P2P服务端，%s", C.Pairname)
			return
		}
	}
}
