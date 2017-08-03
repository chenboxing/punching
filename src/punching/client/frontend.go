package client

import (
	"net"
	. "punching/constant"
	"punching/logger"
	"punching/util"
	"sync"
	"time"
)

var ListenAcceptMap map[string]net.Conn
var ExitChanMap map[string]chan bool

var RWLock *sync.RWMutex

func handleClientConn(source net.Conn) {

	// 4 bits unique session id
	var sessionID string
	// Unique session id
	for {
		RWLock.Lock()
		sessionID = util.GenerateRandomPairKey()
		_, ok := ListenAcceptMap[sessionID]
		RWLock.Unlock()
		if !ok {
			break
		}
	}
	logger.Infof("Enter handleClientConn:%s", sessionID)

	RWLock.Lock()
	ListenAcceptMap[sessionID] = source
	ExitChanMap[sessionID] = make(chan bool)
	RWLock.Unlock()
	logger.Infof("建立Map,%s", sessionID)

	defer func() {

		e1 := source.Close()
		if e1 != nil {
			logger.Error("关闭Sourcer失败")
		}
		RWLock.Lock()
		delete(ListenAcceptMap, sessionID)
		delete(ExitChanMap, sessionID)
		logger.Infof("删除map:%s", sessionID)
		RWLock.Unlock()

	}()

	go func() {

		buf := make([]byte, 1024)
		var flag int

		for {

			len01, err := source.Read(buf)

			if len01 <= 0 || err != nil {
				logger.Errorf("读取Source源连接出错，原因为：%s", err)

				//发送控制
				packQuit := util.PackageNat(PAIR_CONTROL_QUIT, sessionID, []byte(""))
				Wch <- packQuit
				return
			}

			controlID := PAIR_CONTROL_NORMAL
			if flag == 0 {
				// 第一次
				controlID = PAIR_CONTROL_FIRST
				flag = 1
			}

			pack := util.PackageNat(controlID, sessionID, buf[0:len01])
			Wch <- pack

		}

	}()

	select {
	case <-ExitChanMap[sessionID]:
		logger.Warn("需要退出Accept")
		return
	}
}

// 侦听端口，将连接转到natConn
func ClientListenHandle() {

	ListenAcceptMap = make(map[string]net.Conn)
	ExitChanMap = make(map[string]chan bool)
	RWLock = new(sync.RWMutex)

	addrOn := Config.Listen

	l, err := net.Listen("tcp", addrOn)
	if err != nil {
		logger.Errorf("listen %s encountered errors %s", addrOn, err)
		return
	}
	logger.Infof("server running at port %s", addrOn)

	// 全局读取来自nat源的包
	go handleReadConn()

	go func() {
		for {
			c, err := l.Accept()
			if err != nil {
				logger.Errorf("accept error: %s", err)
				break
			}
			go handleClientConn(c)
		}
	}()
}

// 读取目标流到源
func handleReadConn() {
	for {
		select {
		case pact := <-Rch:

			// 获取src
			controlID := pact.ControlID
			sessionID := string(pact.SessionID)
			data := pact.Data

			//退出
			if controlID == PAIR_CONTROL_QUIT {
				if c, ok := ExitChanMap[sessionID]; ok {
					logger.Info("发送退出信号")
					c <- true
				} else {
					logger.Info("在ExitChanMap里找不到Key为:", sessionID)
				}
			} else {
				if src, ok := ListenAcceptMap[sessionID]; ok {
					len2, err2 := src.Write(data)
					if err2 != nil || len2 <= 0 {
						logger.Infof("源写入出错:%s", err2)
					} else {
						logger.Info(time.Now().UnixNano(), "源写入:", len2)
					}

				} else {
					logger.Info("在Map里找不到Key为:", sessionID)
				}

			}
		}
	}
}
