package server

import (
	"log"
	"net"
	"os"
	"sync"
	"punching/util"
	. "punching/constant"
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
			sessionID := string(pack.SessionID)
			data := pack.Data

			//log.Println("读取Nat接收包：handleReadConn", string(r[0:34]), "长度为", len(r))

			if controlID == PAIR_CONTROL_QUIT {

				RWLock.RLock()

				if c, ok := ExitChanMap[sessionID]; ok {
					log.Println("发送退出信号")
					c <- true
				} else {
					log.Println("在ExitChanMap里找不到Key为:", sessionID)
				}
				RWLock.RUnlock()
				break
			}

			//第一次
			if controlID == PAIR_CONTROL_FIRST  {
				log.Println("准备连接:", targetAddr)
				target, err := net.Dial("tcp", targetAddr)
				if err != nil {
					log.Println("连接目标出错", targetAddr)
					break
				}

				ExitChanMap[sessionID] = make(chan bool)
				DialTargetMap[sessionID] = target

				log.Println("连接目标成功:", targetAddr)

				_, err2 := target.Write(pack)
				if err2 != nil {
					log.Println("连接成功后写目标出错", err2.Error())
					break
				}
				go ReadFromTarget(target, sessionID)
			} else {

				if dialtarget, ok := DialTargetMap[sessionID]; ok {

					len2, err2 := dialtarget.Write(data)
					log.Println("已写入:", len2)
					if err2 != nil {
						log.Println("写目标出错", targetAddr, err2.Error())

						//发送控制
						quitPack := util.PackageNat(PAIR_CONTROL_QUIT, [4]byte(sessionID),[]byte(""))
						Wch <- quitPack

						break
					}

				} else {
					log.Println("找不到目标Dial:")
				}

			}

		case <-Dch:
			//出错
			os.Exit(3)
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
				log.Println("读取目标连接数据出错，原因为:", err.Error())

				pack := util.PackageNat(PAIR_CONTROL_QUIT, [4]byte(sessionID),[]byte(""))
				Wch <- pack

				return
			}

			pack := util.PackageNat(PAIR_CONTROL_NORMAL,[4]byte(sessionID), buf[0:j])

			Wch <- pack

		}
	}()

	//接受到退出标识
	select {
	case <-ExitChanMap[sessionID]:
		log.Println("需要退出Accept")
		return
	}

}
