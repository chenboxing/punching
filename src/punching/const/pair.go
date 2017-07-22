package const

// Constant for the client and server

const (
	PAIR_PACT_HEAD     byte = 'P'    // C<->S 自定义包头
	PAIR_CONTROL_FIRST byte = 11     // 控制码 C->S第一个包
	PAIR_CONTROL_QUIT  byte = 10     // 控制码 退出
	PAIR_CONTROL_NORMAL byte = 0     // 控制码  

	CLIENT_PAIR_ACK  byte = 1        // 客户端匹配确认
    SERVER_PAIR_ACK  byte = 2        // 服务端匹配确认	
)

const (
	ROLE_SERVER   int = 1         // 点对点服务端
	ROLE_CLIENT   int = 2         // 点对点客户端
}

