package proxy 

import (
	"net"
	"sync"
	"fmt"
	"punching/util"
	"time"
	. "punching/constant"
	"log"
)

// ServerConn 服务端到代理端连接
type ServerConn struct {
	Rch chan []byte         // 读通道
	Wch chan []byte         // 写通道
	Dch chan bool           // 连接退
	LocalAddr string        // 客户端IP信息 
	Pairname string         // 匹配名称
	SyncAt int64        // 上次心跳时间 
}

// ClientConn 客户端到代理端连接
type ClientConn struct{
	Pairname string        //匹配码
}

// 全局变量
var (
	OnlineServerList  map[string]*ServerConn   // 服务端连接列表Map
	OnlineClientList  map[string]string   // 客户端连接列表Map
	RWLockClient  *sync.RWMutex                //读写锁
	RWLockServer  *sync.RWMutex
)

func Main(){

	// 加载配置信息
	if err := InitConfig(); err != nil{
		log.Println("加载配置信息出错，原因为:%s", err)
		return 
	}

	OnlineServerList = make(map[string]*ServerConn)
	OnlineClientList = make(map[string]string)

	RWLockClient = new(sync.RWMutex)
	RWLockServer = new(sync.RWMutex)

	listenAddr := Config.Listen
	tcpAddr, err := net.ResolveTCPAddr("tcp", listenAddr)
	if err != nil {
		panic(err)
	}

	listen, err := net.ListenTCP("tcp", tcpAddr)

	if err != nil {
		fmt.Println("监听端口失败:", err.Error())
		return
	}
	fmt.Println("已初始化连接，等待客户端连接...")

	for {
		conn, err := listen.AcceptTCP()
		if err != nil {
			fmt.Println("连接异常:", err.Error())
			continue
		}
		fmt.Println("本地地址:", conn.LocalAddr().String(), "来自远程地址", conn.RemoteAddr().String())
		go Handler(conn)
	}
}

// processRoleClient 处理客户端连接
func processRoleClient(conn net.Conn, pairName string ){

	// 判断匹配的服务端是否已经注册
	RWLockServer.RLock()	 
	serverConn,ok := OnlineServerList[pairName]
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

	if !ok{
		packErr := util.PackageProxy(PROXY_CONTROL_ERROR_CLIENT_EXIST,[]byte(""))
		conn.Write(packErr)
		return 
	}

	// 添加到客户端列表
	RWLockClient.Lock()
	OnlineClientList[pairName] = pairName
	RWLockClient.Unlock()

	// 发送Nat地址和接收确认
	toClientAddrs :=  serverConn.LocalAddr + "," + conn.LocalAddr().String()

	pack := util.PackageProxy(PROXY_CONTROL_NORMAL, []byte(toClientAddrs))
	conn.Write(pack)

	buf := make([]byte, 512)
	lenAck, err := conn.Read(buf)

	if err != nil {
		fmt.Println("读客户端确认数据出错")
		return
	}

	ackPack := util.UnpackageProxy(buf[0:lenAck])
	flag := 0
    if ackPack.CotnrolID == PROXY_CONTROL_HEARTBITACK {
		flag += 1
	}
	
	toServerAddrs :=  conn.LocalAddr().String() + "," + serverConn.LocalAddr
	addrPack := util.PackageProxy(PROXY_CONTROL_NORMAL, []byte(toServerAddrs))
	serverConn.Wch <-  addrPack

	// 等待服务端的确认数据
	select {
		case bufAck :=  <- serverConn.Rch:
			pack := util.UnpackageProxy(bufAck)
			if pack.CotnrolID == PROXY_CONTROL_HEARTBITACK {
				flag += 1
			}
			break
	}

	// 收到服务端的确认数据
	if flag == 2 {
		RWLockServer.Lock()
		serverConn.Dch <- true   // 关闭服务端连接
		delete( OnlineServerList, pairName)
		RWLockServer.Unlock()
	}


	RWLockClient.Lock()
	delete( OnlineClientList, pairName)
	RWLockClient.Unlock()
		
	return 

}

// Handle 连接处理函数
func Handler(conn net.Conn) {

	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("连接出现问题:%s",r)
		}
	}()

	defer conn.Close()
	buf := make([]byte, 1024)

	var pairName string
	var C *ServerConn

	// 确定连接类型，判断是否是有效的连接,
	// 对于客户端，需满足
	//   1. 存在对应的服务端
	//   2. 不能存在多个客户端
	// 对于服务端:
	//   1. 
	
	
	for {
		i, err := conn.Read(buf)
		if err != nil {
			fmt.Println("读取数据错误:", err.Error())
			return 
		}

		firstPack := util.UnpackageProxy(buf[0:i])

		clientType := firstPack.Data[0]
		var pairName string
		if len(firstPack.Data) >1 {
			pairName = string(firstPack.Data[1:])
		}

		// 处理客户端连接
		if clientType == ROLE_CLIENT{
			processRoleClient(conn, pairName)
			return  // 退出客户端连接			
		}
		break
	}

	// 下面的操作都是针对服务端连接

	// 服务端连接允许匹配码为空，系统将随机产生唯一匹配码
	if pairName == ""{
		for{
			pairName = util.GenerateRandomPairKey()
			RWLockServer.Lock()
			if _, ok := OnlineServerList[pairName]; !ok {
				break;
			}
			RWLockServer.Unlock()
		}
	}else{

		// 是否存在pair name
		RWLockServer.RLock()
		_, ok := OnlineServerList[pairName]
		RWLockServer.RUnlock()

		// 已存在
		if ok {
			errPack := util.PackageProxy(PROXY_CONTROL_ERROR_SERVER_EXIST, []byte(""))
			conn.Write(errPack)
			fmt.Printf("服务端列表中已存在:%s", pairName)
			return
		}

	}


	// 生成服务器连接对象添加到列表
	RWLockServer.Lock()
	serverConn := &ServerConn{Rch: make(chan []byte), Wch: make(chan []byte), Pairname: pairName, LocalAddr: conn.LocalAddr().String()}
	OnlineServerList[pairName] = serverConn
	RWLockServer.Unlock()

	//	写通道
	go WHandler(conn, serverConn)

	//	读通道
	go RHandler(conn, serverConn)
	
	// 等待退出通道
	select {
	case <-C.Dch:
		fmt.Println("close handler goroutine")
	}
}

// 正常写数据 匹配端连接上来会写信息
// 定时检测 conn die => goroutine die
func WHandler(conn net.Conn, C *ServerConn) {
	// 读取业务Work 写入Wch的数据
	ticker := time.NewTicker(60 * time.Second)
	for {
		select {
		case d := <-C.Wch:
			conn.Write(d)
		case <-ticker.C:  //60秒无操作,可能连接已中断
			RWLockServer.RLock()
			_, ok := OnlineServerList[C.Pairname];
			RWLockServer.RUnlock()				
			if !ok {
				fmt.Println("conn die, close WHandler")
				return
			}
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
		err := conn.SetReadDeadline(time.Now().Add(10 * time.Second))
		if err != nil {
			fmt.Println(err)
		}
		if i, derr := conn.Read(data); derr == nil {
			// 可能是来自客户端的消息确认
			//           	     数据消息
			pack := util.UnpackageProxy(data[0:i])
			if pack.CotnrolID == PROXY_CONTROL_HEARTBITACK {
				fmt.Println("Received hartbeat ack")//// C.Rch <- data
			}

			continue
		}

		//如果等待10秒没有读到客户端数据或读写出错，写心跳包
		// 写心跳包
		heartPack := util.PackageProxy(PROXY_CONTROL_HEARTBIT, []byte(""))
		conn.Write(heartPack)

		conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		if _, herr := conn.Read(data); herr == nil {
			
			// fmt.Println(string(data))
			// 更新心跳时间
			RWLockServer.RLock()
			serverConn, ok := OnlineServerList[C.Pairname];
			if ok {
				serverConn.SyncAt = time.Now().Unix()
			}
			RWLockServer.RUnlock()

			fmt.Println("resv ht packet ack")
		} else {
			RWLockServer.Lock()
			delete(OnlineServerList, C.Pairname)
			RWLockServer.Unlock()
			fmt.Println("delete user!")
			return
		}
	}
}


