package proxy 

import (
	"net"
	"sync"
	"fmt"
	"punching/util"
	"punching/logger"
	"time"
	"punching/const"
)

const (	
	PROXY_FIRST_HEAD byte = 'P'     //包头标识
	ROLE_SERVER byte =  1           //服务端
	ROLE_CLIENT byte =  2           //客户端
	PROXY_RESP_NO_SERVER  = "no_server_partner"  // 代理响应-服务端还没有启用
	PROXY_RESP_CLIENT_EXIST  = "client_existed"  // 代理响应-客户端已经存在了
	PROXY_RESP_ERR_SERVER_EXIST  = "server_existed"  // 代理响应-客户端已经存在了
	
	CLIENT_RESP_ACK  byte = 1 
    SERVER_RESP_ACK  byte = 1	
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
	OnlineServerList  map[string]ServerConn   // 服务端连接列表Map
	OnlineClientList  map[string]ClientConn   // 客户端连接列表Map
	RWLockClient  sync.RWMutex                //读写锁
	RWLockServer  sync.RWMutex
)

func Main(){

	// 加载配置信息
	if err := InitConfig(); err != nil{
		log.Println("加载配置信息出错，原因为:%s", err)
		return 
	}


	OnlineServerList = make(map[string]ServerConn)
	OnlineClientList = make(map[string]ClientConn)

	RWLockClient = new(sync.RWMutex)
	RWLockServer = new(sync.RWMutex)

	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
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

// parseFirstPackage 解析连接的第一个条
// [1]+[1]+[4-32] (包头标识+类型+32字节的Pairname)
func parseFirstPackage(data []byte)(pairname string, roleType int, err error){
   if len(data) <= 6 || len(data) > 34 {
	   err = fmt.Errorf("%s","包长度不匹配")
	   return 
   }

   roleType = int(data[1])
   if data[0] != PROXY_FIRST_HEAD ||  ( roleType != ROLE_CLIENT && roleType != ROLE_SERVER) {
	   err = fmt.Errorf("%s","包内容不匹配")
	   return 
   }
      
   pairname = string(byte[2:]) 
   
   return 
}

// processRoleClient 处理客户端连接
func processRoleClient(conn net.Conn，pairname string ){

	// 判断匹配的服务端是否已经注册
	RWLockServer.RLock()	 
	serverConn,ok := OnlineServerList[pairname]
	RWLockServer.RUnlock()

	if !ok {
		// 客户端没有注册
		conn.Write([]byte(PROXY_RESP_NO_SERVER))
		return 
	}

	RWLockClient.RLock()
	ok, _ := OnlineClientList[pairname]
	RWLockClient.RUnlock()

	if !ok{
		conn.Write([]byte(PROXY_RESP_CLIENT_EXIST))
		return 
	}

	// 添加到客户端列表
	RWLockClient.Lock()
	OnlineClientList[pairname] = pairname
	RWLockClient.Unlock()
	
	// byteNat := bytes.NewBuffer(nil)
 	// byteNat.write([]byte())
	//  result.Bytes()
	 
	
	// 发送Nat地址和接收确认
	toClientAddrs :=  serverConn.LocalAddr + "," + conn.LocalAddr	
	conn.Write(byte[]{toClientAddrs})

	buf := make([]byte, 128)
	lenAck, err := conn.Read(buf)

	flag := 0 

    if buf[0] == CLIENT_RESP_ACK {
		flag += 1
	}
	
	toServerAddrs :=  conn.LocalAddr + "," + serverConn.LocalAddr
	serverSide.Wch <- toServerAddrs

	// 等待服务端的确认数据
	select {
		case bufAck :=  <- serverSide.Rch
		if bufAck[0] == SERVER_RESP_ACK {
			flag += 1 
			break 
		}				
	}

	// 收到服务端的确认数据
	if flag == 2 {
		RWLockServer.Lock()
		serverConn.Dch <- true   // 关闭服务端连接
		delete( OnlineServerList, pairname)
		RWLockServer.Unlock()
	}


	RWLockClient.Lock()
	delete( OnlineClientList, pairname)
	RWLockClient.Unlock()
		
	return 

}

// Handle 连接处理函数
func Handler(conn net.Conn) {

	defer func() {
		if r := recover(); r != nil {
			logger.Println("连接出现问题:%s",r.Error())
		}
	}()	

	defer conn.Close()
	
	buf := make([]byte, 1024)
	var uid string
	var C *OnlineServerSide

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

		//获取匹配名称和连接类型（服务端或客户端)
		pairname, roleType, err := parseFirstPackage(buf[0:i])
		if err != nil {
			return 
		}
		
		// 处理客户端连接
		if roleType == ROLE_CLIENT{
			processRoleClient(conn)
			return  // 退出客户端连接			
		}
		
		
		
			
			
			break 		
			
		
	}


	// 下面的操作都是针对服务端连接
			// 是否存在pair name
	RWLockServer.RLock()			 
	_, ok := OnlineServerList[pairname]
	RWLockServer.RUnlock()

	// 已存在
	if ok {
		conn.Write([]byte(PROXY_RESP_ERR_SERVER_EXIST))	
		return 	
	}

	// 生成服务器连接对象添加到列表
	RWLockServer.Lock()
	serverConn := &ServerConn{Rch: make(chan []byte), Wch: make(chan []byte), Pairname: pairname, LocalAddr: conn.LocalAddr}
	OnlineServerList[pairname] = serverConn
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
func RHandler(conn net.Conn, C *ServerSide) {
	// 心跳ack
	// 业务数据 写入Wch

	for {
		data := make([]byte, 128)
		// 设置读超时
		err := conn.SetReadDeadline(time.Now().Add(10 * time.Second))
		if err != nil {
			fmt.Println(err)
		}
		if _, derr := conn.Read(data); derr == nil {
			// 可能是来自客户端的消息确认
			//           	     数据消息
			fmt.Println(data)
			if data[0] == Res {
				fmt.Println("recv client data ack")
			} else if data[0] == Req {
				fmt.Println("recv client data")
				fmt.Println(data)
				conn.Write([]byte{Res, '#'})
				// C.Rch <- data
			}

			continue
		}

		//如果等待10秒没有读到客户端数据或读写出错，写心跳包

		// 写心跳包
		conn.Write([]byte{Req_HEARTBEAT, '#'})
		fmt.Println("send ht packet")
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
			delete(OnlineServerSide, C.Pairname)
			RWLockServer.Unlock()
			fmt.Println("delete user!")
			return
		}
	}
}


