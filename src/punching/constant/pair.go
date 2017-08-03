package constant

// Constant for the client and server

const (
	PAIR_CONTROL_FIRST  byte = 11 // 控制码 C->S第一个包
	PAIR_CONTROL_QUIT   byte = 10 // 控制码 退出
	PAIR_CONTROL_NORMAL byte = 0  // 控制码

	PAIR_PACKAGE_HEAD_LENGTH      = 6  // C<->S 自定义包头长度
	PAIR_PACKAGE_CONTROL_LENGTH   = 1  // 包控制码长度
	PAIR_PACKAGE_SESSIONID_LENGTH = 4  // 包会话ID长度
	PAIR_PACKAGE_DATA_LENGTH      = 4  // 包数据长度
	PAIR_PACKAGE_PREFIX_LENGTH    = 15 // head[6] + control[1] + sessionid[4] + data length[4]

)

const (
	ROLE_SERVER byte = 1 // 点对点服务端
	ROLE_CLIENT byte = 2 // 点对点客户端
)

var (
	PAIR_PACKAGE_HEAD string = "CBXNAT" // C<->S 自定义包头

)
