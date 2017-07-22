package client

import (
	"log"
	"net"
	"os"
	"reflect"
	"sync"
	"unsafe"
)

var DialTargetMap map[string]net.Conn

func b2s(buf []byte) string {
	return *(*string)(unsafe.Pointer(&buf))
}

func s2b(s *string) []byte {
	return *(*[]byte)(unsafe.Pointer((*reflect.SliceHeader)(unsafe.Pointer(s))))
}

func handleServerConn() {

	DialTargetMap = make(map[string]net.Conn)
	ExitChanMap = make(map[string]chan bool)
	RWLock = new(sync.RWMutex)

	//var target net.Conn
	//var err error
	// defer func() {
	// 	if target != nil {
	// 		target.Close()
	// 	}
	// 	//source.Close()
	// }()

	var targetIP string
	targetIP = "192.168.3.2:5901"
	for {
		select {
		case r := <-Rch:

			//确定target是否存在,如果不存在，重新生成target

			//分析数据包
			log.Println("接收到数据:", len(r.Data))

			controlid := r.ControlID
			uniqueid := r.Key
			pack := r.Data

			//log.Println("读取Nat接收包：handleReadConn", string(r[0:34]), "长度为", len(r))

			if controlid == "01" {

				RWLock.RLock()

				if c, ok := ExitChanMap[uniqueid]; ok {
					log.Println("发送退出信号")
					c <- true
				} else {
					log.Println("在ExitChanMap里找不到Key为:", uniqueid)
				}
				RWLock.RUnlock()
				break
			}

			//第一次
			if controlid == "11" {
				log.Println("准备连接:", targetIP)
				target, err := net.Dial("tcp", targetIP)
				if err != nil {
					log.Println("连接目标出错", targetIP)
					break
				}

				ExitChanMap[uniqueid] = make(chan bool)
				DialTargetMap[uniqueid] = target

				log.Println("连接目标成功:", targetIP)

				_, err2 := target.Write(pack)
				if err2 != nil {
					log.Println("连接成功后写目标出错", err2.Error())
					break
				}
				go ReadFromTarget(target, uniqueid)
			} else {

				if dialtarget, ok := DialTargetMap[uniqueid]; ok {

					len2, err2 := dialtarget.Write(pack)
					log.Println("已写入:", len2)
					if err2 != nil {
						log.Println("写目标出错", targetIP, err2.Error())
						//发送控制
						pack01 := Packet(uniqueid, "01", []byte(""))
						Wch <- pack01

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
func ReadFromTarget(target net.Conn, uniqueid string) {
	defer func() {
		target.Close()

		RWLock.Lock()
		delete(DialTargetMap, uniqueid)
		delete(ExitChanMap, uniqueid)
		RWLock.Unlock()
	}()

	go func() {
		//buf := make([]byte, 512-34)
		buf := make([]byte, 1024)
		for {

			j, err := target.Read(buf)

			if err != nil || j == 0 {
				log.Println("读取目标连接数据出错，原因为:", err.Error())

				pack := Packet(uniqueid, "01", []byte(""))
				Wch <- pack

				return
			}

			pack := Packet(uniqueid, "00", buf[0:j])

			Wch <- pack

		}
	}()

	//接受到退出标识
	select {
	case <-ExitChanMap[uniqueid]:
		log.Println("需要退出Accept")
		return
	}

}
