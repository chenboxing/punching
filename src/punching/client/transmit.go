package client

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sync"
	"time"
)

var ListenAcceptMap map[string]net.Conn
var ExitChanMap map[string]chan bool

var RWLock *sync.RWMutex

//生成32位md5字串
func GetMd5String(s string) string {
	h := md5.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

//生成Guid字串
func UniqueId() string {
	b := make([]byte, 48)

	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return ""
	}
	return GetMd5String(base64.URLEncoding.EncodeToString(b))
}

func handleClientConn(source net.Conn) {

	// 32位唯一码
	uniqueid := UniqueId()
	log.Println("Enter handleClientConn:", uniqueid)

	RWLock.Lock()
	ListenAcceptMap[uniqueid] = source
	ExitChanMap[uniqueid] = make(chan bool)
	RWLock.Unlock()
	log.Println("建立Map", uniqueid)

	defer func() {

		e1 := source.Close()
		if e1 != nil {
			log.Println("关闭Sourcer失败")
		}
		RWLock.Lock()
		delete(ListenAcceptMap, uniqueid)
		delete(ExitChanMap, uniqueid)
		log.Println("删除map", uniqueid)
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
				pack01 := Packet(uniqueid, "01", []byte(""))
				Wch <- pack01
				return
			}

			controlID := "00"
			if flag == 0 {
				// 第一次
				controlID = "11"
				flag = 1
			}
			pack := Packet(uniqueid, controlID, buf[0:len01])
			Wch <- pack

		}

	}()

	select {
	case <-ExitChanMap[uniqueid]:
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

	select {
	case <-Dch:
		os.Exit(3)
	}

}

// 读取目标流到源
func handleReadConn() {
	for {
		select {
		case r := <-Rch:

			log.Println(time.Now().UnixNano(), "handleReadConn准备处理")
			// 获取src
			controlid := r.ControlID
			uniqueid := r.Key
			data := r.Data

			log.Println("读取Nat包：handleReadConn", uniqueid, "长度为", len(data))

			//退出
			if controlid == "01" {
				if c, ok := ExitChanMap[uniqueid]; ok {
					log.Println("发送退出信号")
					c <- true
				} else {
					log.Println("在ExitChanMap里找不到Key为:", uniqueid)
				}
			} else {
				if src, ok := ListenAcceptMap[uniqueid]; ok {
					len2, err2 := src.Write(data)
					if err2 != nil || len2 <= 0 {
						log.Println("源写入出错", err2.Error())
					}
					log.Println(time.Now().UnixNano(), "源写入:", len2)
				} else {
					log.Println("在Map里找不到Key为:", uniqueid)
				}

			}
		}
	}
}
