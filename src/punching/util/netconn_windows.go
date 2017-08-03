package util

import (
	"fmt"
	"log"
	"net"
	"punching/logger"
	"syscall"
	"time"
)

func MAKEWORD(low, high uint8) uint32 {
	var ret uint16 = uint16(high)<<8 + uint16(low)
	return uint32(ret)
}

type NetConn struct {
	sock syscall.Handle
}

func (hole *NetConn) Close() {

	syscall.WSACleanup()
	syscall.Closesocket(hole.sock)

}

func (hole *NetConn) Bind(addr string) (err error) {

	proto := "tcp"

	var wsadata syscall.WSAData

	if err = syscall.WSAStartup(MAKEWORD(2, 2), &wsadata); err != nil {
		log.Println("Startup error")
		return
	}

	var sock syscall.Handle
	syscall.ForkLock.RLock()
	if sock, err = syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, syscall.IPPROTO_TCP); err != nil {
		syscall.ForkLock.RUnlock()
		return
	}
	syscall.ForkLock.RUnlock()

	defer func() {
		if err != nil {
			syscall.Close(sock)
		}
	}()

	if err = syscall.SetsockoptInt(sock, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1); err != nil {
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
		if err = syscall.Bind(sock, sockaddr); err != nil {
			return
		}
	}

	hole.sock = sock
	return
}

func (hole *NetConn) Connect(addr [4]byte, port int) (err error) {
	if hole.sock == 0 {
		err = fmt.Errorf("请先执行Bind()")
		return
	}
	addrInet4 := syscall.SockaddrInet4{
		Addr: addr,
		Port: port,
	}

	chConnect := make(chan error)
	logger.Info(time.Now().UnixNano(), "准备连接对方")
	go func() {
		err = syscall.Connect(hole.sock, &addrInet4)
		chConnect <- err
	}()

	//有时候连接被远端抛弃的时候， syscall.Connect() 会很久才返回
	ticker := time.NewTicker(30 * time.Second)
	select {
	case <-ticker.C:
		err = fmt.Errorf("Connect timeout")
		return
	case e := <-chConnect:
		if e != nil {
			err = e
			log.Println("Connect error: ", err)
			return
		}
	}
	return nil
}

func (hole *NetConn) Read(buffer []byte) (length int, err error) {

	dataWsaBuf := syscall.WSABuf{Len: uint32(len(buffer)), Buf: &buffer[0]}
	flags := uint32(0)
	recvd := uint32(0)

	err = syscall.WSARecv(hole.sock, &dataWsaBuf, 1, &recvd, &flags, nil, nil)
	if err != nil {
		return 0, err
	}
	return int(recvd), nil
}

func (hole *NetConn) Write(data []byte) (length int, err error) {
	var (
		dataWsaBuf syscall.WSABuf
		SendBytes  uint32
		overlapped syscall.Overlapped
	)
	dataWsaBuf.Len = uint32(len(data))
	dataWsaBuf.Buf = &data[0]
	err = syscall.WSASend(hole.sock, &dataWsaBuf, 1, &SendBytes, 0, &overlapped, nil)
	if err != nil {
		return 0, err
	} else {
		return int(SendBytes), nil
	}
}
