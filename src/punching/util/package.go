package util

import (
	"bytes"
	. "punching/constant"
	"log"
)

// PairPackage P2P端通讯封装包
type PairPackage struct {
	Head      [6]byte    // 头
	ControlID byte    // 控制ID
	SessionID [4]byte // 会话ID
	Data      []byte  // 数据
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
func UnpackageProxy(pact []byte) (control byte, data []byte) {
	return pact[1], pact[1:]
}

// Customize P2P data package
// The format of package defined below:
// head(6)+control(1)+session id(4) + data length (4) + data
func PackageNat(control byte, sessionID [4]byte, data []byte) []byte {
	pack := bytes.NewBuffer(nil)
	pack.Write([]byte(PAIR_PACKAGE_HEAD)) // Head [6]byte
	pack.Write([]byte{control})
	pack.Write([]byte(sessionID))
	pack.Write(IntToBytes(len(data)))            // length of sent data
	pack.Write([]byte(sessionID))

	pack.Write(data)
	return pack.Bytes()
}

// Nat网络后面的Client端和Server端拆包
// 需要考虑沾包，分析出的完整封装包传入读channel
func UnpackageNat(buffer []byte, readChan chan PairPackage) (data []byte) {

	length := len(buffer)
	log.Println("长度为:", length)
	var i int
	for i = 0; i < length; i = i + 1 {
		if length < i + PAIR_PACKAGE_PREFIX_LENGTH {
			break
		}

		if string(buffer[i:i+ PAIR_PACKAGE_HEAD_LENGTH]) == string(PAIR_PACKAGE_HEAD) {
			// Length of data
			dataLength := BytesToInt(buffer[i+ PAIR_PACKAGE_PREFIX_LENGTH - PAIR_PACKAGE_DATA_LENGTH: i +
				PAIR_PACKAGE_PREFIX_LENGTH])

			if length < i+ PAIR_PACKAGE_PREFIX_LENGTH + dataLength {
				break
			}

			// data
			data := buffer[i + PAIR_PACKAGE_PREFIX_LENGTH: i +  PAIR_PACKAGE_PREFIX_LENGTH +
				dataLength]

			controlID := buffer[i + PAIR_PACKAGE_HEAD_LENGTH : i + PAIR_PACKAGE_HEAD_LENGTH +
				PAIR_PACKAGE_CONTROL_LENGTH]
			sessionID := string(buffer[i+ PAIR_PACKAGE_HEAD_LENGTH + PAIR_PACKAGE_CONTROL_LENGTH:
				i + PAIR_PACKAGE_HEAD_LENGTH + PAIR_PACKAGE_CONTROL_LENGTH + PAIR_PACKAGE_SESSIONID_LENGTH])

			pact := PairPackage{
				Head:   PAIR_PACKAGE_HEAD,
				Data:      data,
				ControlID: controlID[0],
				SessionID: [4]byte(sessionID),
			}
			readChan <- pact

			i +=  PAIR_PACKAGE_PREFIX_LENGTH + dataLength - 1
		}
	}

	if i == length {
		return make([]byte, 0)
	}
	return buffer[i:]

}
