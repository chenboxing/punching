package util

import (
	"fmt"
	"log"
	"net"
	"os"
	"punching/logger"
	"syscall"
	"time"
)

type NetConn struct {
	fd   int      // 文件句柄
	conn net.Conn // 连接对象
}

func (hole *NetConn) Close() {
	//if hole.conn != nil {
	hole.conn.Close()
	//}

}

func (hole *NetConn) Bind(addr string) (err error) {

	proto := "tcp"

	syscall.ForkLock.RLock()
	var fd int
	if fd, err = syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, syscall.IPPROTO_TCP); err != nil {
		syscall.ForkLock.RUnlock()
		return
	}
	syscall.ForkLock.RUnlock()

	defer func() {
		if err != nil {
			syscall.Close(fd)
		}
	}()

	if err = syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1); err != nil {
		return
	}

	if len(addr) > 0 {
		var tcp *net.TCPAddr
		tcp, err = net.ResolveTCPAddr(proto, addr)
		if err != nil && tcp.IP != nil {
			log.Println(err)
			return
		}
		sockaddr := &syscall.SockaddrInet4{Port: tcp.Port}
		if err = syscall.Bind(fd, sockaddr); err != nil {
			return
		}
	}

	hole.fd = fd

	return
}

func (hole *NetConn) Connect(addr [4]byte, port int) (err error) {

	if hole.fd == 0 {

		err = fmt.Errorf("请先调用Bind()函数")
		return
	}

	addrInet4 := syscall.SockaddrInet4{
		Addr: addr,
		Port: port,
	}

	chConnect := make(chan error)
	logger.Info(time.Now().UnixNano(), "准备连接对方")
	go func() {
		err = syscall.Connect(hole.fd, &addrInet4)
		chConnect <- err
	}()

	//有时候连接被远端抛弃的时候， syscall.Connect() 会很久才返回
	ticker := time.NewTicker(60 * time.Second)
	select {
	case <-ticker.C:
		err = fmt.Errorf("Connect timeout")
		return
	case e := <-chConnect:
		if e != nil {
			err = e
			logger.Errorf("Connect error: %s", err)
			return
		}
	}

	// 转为net.conn对象
	var file *os.File
	file = os.NewFile(uintptr(hole.fd), fmt.Sprintf("tcpholepunching.%d", time.Now().UnixNano()))
	if conn0, err0 := net.FileConn(file); err0 != nil {
		log.Println("Connect error", err0)
		err = err0
		return
	} else {
		hole.conn = conn0
	}

	if err = file.Close(); err != nil {
		log.Println("Connect error", err)
		return
	}
	return

}

func (hole *NetConn) Read(buffer []byte) (length int, err error) {

	return hole.conn.Read(buffer)
}

func (hole *NetConn) Write(data []byte) (length int, err error) {

	return hole.conn.Write(data)
}
