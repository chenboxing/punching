package client

import (
	"fmt"
	"log"
	"net"
	. "punching/constant"
	"sync"
	"time"
	"punching/util"
)

var ListenAcceptMap map[string]net.Conn
var ExitChanMap map[string]chan bool

var RWLock *sync.RWMutex


func handleClientConn(source net.Conn) {

	// 4 bits unique session id
	var sessionID string
	for{
		RWLock.Lock()
		sessionID = util.GenerateRandomPairKey()
		if _, ok := ListenAcceptMap[sessionID]; !ok{
			break
		}
		RWLock.Unlock()
	}
	log.Println("Enter handleClientConn:", sessionID)

	RWLock.Lock()
	ListenAcceptMap[sessionID] = source
	ExitChanMap[sessionID] = make(chan bool)
	RWLock.Unlock()
	log.Println("建立Map", sessionID)

	defer func() {

		e1 := source.Close()
		if e1 != nil {
			log.Println("关闭Sourcer失败")
		}
		RWLock.Lock()
		delete(ListenAcceptMap, sessionID)
		delete(ExitChanMap, sessionID)
		log.Println("删除map", sessionID)
		RWLock.Unlock()

	}()

	go func() {

		buf := make([]byte, 1024)
		var flag int

		for {

			len01, err := source.Read(buf)

			if len01 <= 0 || err != nil {
				log.Println("读取Source源连接出错，原因为：", err.Error())

				//发送控制
				packQuit := util.PackageNat(PAIR_CONTROL_QUIT, [4]byte(sessionID),[]byte("") )
				Wch <- packQuit
				return
			}

			controlID :=  PAIR_CONTROL_NORMAL
			if flag == 0 {
				// 第一次
				controlID = PAIR_CONTROL_FIRST
				flag = 1
			}
			pack :=  util.PackageNat(controlID, [4]byte(sessionID), buf[0:len01])
			Wch <- pack

		}

	}()

	select {
	case <-ExitChanMap[sessionID]:
		log.Println("需要退出Accept")
		return
	}
}

// 侦听端口，将连接转到natConn
func ClientListenHandle() {

	ListenAcceptMap = make(map[string]net.Conn)
	ExitChanMap = make(map[string]chan bool)
	RWLock = new(sync.RWMutex)

	addrOn := Config.Dial

	l, err := net.Listen("tcp", addrOn)
	if err != nil {
		fmt.Println("listen ", addrOn, " error:", err)
		return
	}
	fmt.Println("server running at port", addrOn)

	// 全局读取来自nat源的包
	go handleReadConn()

	go func() {
		for {
			c, err := l.Accept()
			if err != nil {
				fmt.Println("accept error:", err)
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

			log.Println(time.Now().UnixNano(), "handleReadConn准备处理")
			// 获取src
			controlID := pact.ControlID
			sessionID := string(pact.SessionID)
			data := pact.Data

			log.Println("读取Nat包：handleReadConn", sessionID, "长度为", len(data))

			//退出
			if controlID == PAIR_CONTROL_QUIT {
				if c, ok := ExitChanMap[sessionID]; ok {
					log.Println("发送退出信号")
					c <- true
				} else {
					log.Println("在ExitChanMap里找不到Key为:", sessionID)
				}
			} else {
				if src, ok := ListenAcceptMap[sessionID]; ok {
					len2, err2 := src.Write(data)
					if err2 != nil || len2 <= 0 {
						log.Println("源写入出错", err2.Error())
					}
					log.Println(time.Now().UnixNano(), "源写入:", len2)
				} else {
					log.Println("在Map里找不到Key为:", sessionID)
				}

			}
		}
	}
}
