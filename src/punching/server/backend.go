package server

import (
	"net"
	"os"
	. "punching/constant"
	"punching/logger"
	"punching/util"
	"sync"
)

var ExitChanMap map[string]chan bool
var RWLock *sync.RWMutex
var DialTargetMap map[string]net.Conn

func handleServerConn() {

	DialTargetMap = make(map[string]net.Conn)
	ExitChanMap = make(map[string]chan bool)
	RWLock = new(sync.RWMutex)

	var targetAddr = Config.Dial
	for {
		select {
		case pack := <-Rch:

			//确定target是否存在,如果不存在，重新生成target

			controlID := pack.ControlID
			sessionID := pack.SessionID
			data := pack.Data

			//log.Println("读取Nat接收包：handleReadConn", string(r[0:34]), "长度为", len(r))

			if controlID == PAIR_CONTROL_QUIT {

				RWLock.RLock()

				if c, ok := ExitChanMap[sessionID]; ok {
					logger.Info("发送退出信号")
					c <- true
				} else {
					logger.Errorf("在ExitChanMap里找不到Key为:%s", sessionID)
				}
				RWLock.RUnlock()
				break
			}

			//第一次
			if controlID == PAIR_CONTROL_FIRST {
				logger.Info("准备连接:", targetAddr)
				target, err := net.Dial("tcp", targetAddr)
				if err != nil {
					logger.Errorf("连接目标出错:%s", targetAddr)
					break
				}

				ExitChanMap[sessionID] = make(chan bool)
				DialTargetMap[sessionID] = target

				_, err2 := target.Write(pack.Data)
				if err2 != nil {
					logger.Errorf("连接成功后写目标出错,%s", err2)
					break
				}
				go ReadFromTarget(target, sessionID)
			} else {

				if dialtarget, ok := DialTargetMap[sessionID]; ok {

					len2, err2 := dialtarget.Write(data)
					logger.Info("已写入:", len2)
					if err2 != nil {
						logger.Errorf("写目标:%s,出错:%s", targetAddr, err2)

						//发送控制
						quitPack := util.PackageNat(PAIR_CONTROL_QUIT, sessionID, []byte(""))
						Wch <- quitPack

						break
					}

				} else {
					logger.Errorf("找不到目标Dial:%s", sessionID)
				}

			}

		case <-Dch:
			//出错
			logger.Warn("收到退出信息")
			os.Exit(1)
		}
	}
}

// 读取目标流到源
func ReadFromTarget(target net.Conn, sessionID string) {

	defer func() {
		target.Close()
		RWLock.Lock()
		delete(DialTargetMap, sessionID)
		delete(ExitChanMap, sessionID)
		RWLock.Unlock()
	}()

	go func() {
		//buf := make([]byte, 512-34)
		buf := make([]byte, 1024)
		for {

			j, err := target.Read(buf)

			if err != nil || j == 0 {
				logger.Errorf("读取目标连接数据出错，原因为:%s", err)

				pack := util.PackageNat(PAIR_CONTROL_QUIT, sessionID, []byte(""))
				Wch <- pack
				return
			}
			logger.Info("准备构造数据")
			pack := util.PackageNat(PAIR_CONTROL_NORMAL, sessionID, buf[0:j])

			Wch <- pack

		}
	}()

	//接受到退出标识
	select {
	case <-ExitChanMap[sessionID]:
		logger.Warn("需要退出Accept")
		return
	}

}
