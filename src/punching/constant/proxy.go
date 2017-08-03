package constant

const (
	PROXY_PACKAGE_HEAD               byte = 'H' // C<->S 自定义包头
	PROXY_CONTROL_FIRST              byte = 11  // 控制ID 第一个数据包
	PROXY_CONTROL_NORMAL             byte = 0   // 控制码 正常发送
	PROXY_CONTROL_ACK                byte = 12  // 控制码 确认
	PROXY_CONTROL_QUIT               byte = 10  // 控制码 退出
	PROXY_CONTROL_HEARTBIT           byte = 13  // 控制码 心跳包
	PROXY_CONTROL_HEARTBITACK        byte = 14  //  心跳包确认
	PROXY_CONTROL_ERROR_NO_SERVER    byte = 201 // 服务端还没有注册
	PROXY_CONTROL_ERROR_CLIENT_EXIST byte = 202 // 客户端已经存在
	PROXY_CONTROL_ERROR_SERVER_EXIST byte = 203 // 服务端已经存在

)
