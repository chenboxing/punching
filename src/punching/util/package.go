package util

import (
	"bytes"
	"errors"
	. "punching/constant"
)

// PairPackage P2P端通讯封装包
type PairPackage struct {
	Head      string // 头 6位字符
	ControlID byte   // 控制ID
	SessionID string // 会话ID 4位字符
	Data      []byte // 数据
}

// 代理解析端通讯包
type ProxyPackage struct {
	Head      byte   // 头
	ControlID byte   // 控制ID
	Data      []byte // 数据
}

// 跟代理解析端通讯封包
func PackageProxy(control byte, data []byte) []byte {
	pack := bytes.NewBuffer(nil)
	pack.Write([]byte{PROXY_PACKAGE_HEAD})
	pack.Write([]byte{control})
	pack.Write(data)
	return pack.Bytes()
}

// 跟代理解析端拆包
func UnpackageProxy(buffer []byte) (pack ProxyPackage, err error) {

	if len(buffer) < 2 {
		err = errors.New("格式不对，长度小于2")
		return
	}

	if buffer[0] != PROXY_PACKAGE_HEAD {
		err = errors.New("包头不对")
		return
	}
	pack = ProxyPackage{buffer[0], buffer[1], buffer[2:]}
	return
}

// Customize P2P data package
// The format of package defined below:
// head(6)+control(1)+session id(4) + data length (4) + data
func PackageNat(control byte, sessionID string, data []byte) []byte {
	pack := bytes.NewBuffer(nil)
	pack.Write([]byte(PAIR_PACKAGE_HEAD)) // Head [6]byte
	pack.Write([]byte{control})
	pack.Write([]byte(sessionID))
	pack.Write(IntToBytes(len(data))) // length of sent data

	pack.Write(data)
	return pack.Bytes()
}

// Nat网络后面的Client端和Server端拆包
// 需要考虑沾包，分析出的完整封装包传入读channel
func UnpackageNat(buffer []byte, readChan chan PairPackage) (data []byte) {

	length := len(buffer)

	var i int
	for i = 0; i < length; i = i + 1 {
		if length < i+PAIR_PACKAGE_PREFIX_LENGTH {
			break
		}

		if string(buffer[i:i+PAIR_PACKAGE_HEAD_LENGTH]) == PAIR_PACKAGE_HEAD {

			// Length of data
			dataLength := BytesToInt(buffer[i+PAIR_PACKAGE_PREFIX_LENGTH-PAIR_PACKAGE_DATA_LENGTH : i+
				PAIR_PACKAGE_PREFIX_LENGTH])

			if length < i+PAIR_PACKAGE_PREFIX_LENGTH+dataLength {
				break
			}

			// data
			data := buffer[i+PAIR_PACKAGE_PREFIX_LENGTH : i+PAIR_PACKAGE_PREFIX_LENGTH+
				dataLength]

			controlID := buffer[i+PAIR_PACKAGE_HEAD_LENGTH : i+PAIR_PACKAGE_HEAD_LENGTH+
				PAIR_PACKAGE_CONTROL_LENGTH]

			iSessionIDStartPos := PAIR_PACKAGE_HEAD_LENGTH + PAIR_PACKAGE_CONTROL_LENGTH
			sessionID := string(buffer[i+iSessionIDStartPos : i+iSessionIDStartPos+PAIR_PACKAGE_SESSIONID_LENGTH])

			pack := PairPackage{
				Head:      PAIR_PACKAGE_HEAD,
				Data:      data,
				ControlID: controlID[0],
				SessionID: sessionID,
			}

			readChan <- pack

			i += PAIR_PACKAGE_PREFIX_LENGTH + dataLength - 1
		}
	}

	if i == length {
		return make([]byte, 0)
	}
	return buffer[i:]

}
