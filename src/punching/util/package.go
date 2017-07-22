package  util 

const {
	PACKAGE_PROXY_HEADER byte = 'P'
	PACKAGE_NAT_HEADER   [6]byte = []byte{'C','B','X','N','A','T'}
}

// 跟代理解析端通讯封包 
func PackageProxy(control byte, data []byte)  []byte{
	pack := bytes.NewBuffer(nil)
	pack.Write(PACKAGE_PROXY_HEADER)
	pack.Write(control)
	pack.Write(data)
	return pack.Bytes()
}

// 跟代理解析端通讯拆包 
func UnpackageProxy(pact []byte)(control byte, data []byte) {

}

// Nat网络后面的Client端和Server端封包
func PackageNat(control byte, sessionKey [4]byte, data []byte) []byte{
	
}

// Nat网络后面的Client端和Server端拆包
// 需要考虑沾包，分析出的完整封装包传入读channel
func UnpackageNat(pact []byte)(data []byte){

	length := len(buffer)
	log.Println("长度为:", length)
	var i int
	for i = 0; i < length; i = i + 1 {
		if length < i+ConstHeaderLength+ConstControlLength+ConstDataLength {
			break
		}
		if string(buffer[i:i+len(ConstHeaderKey)]) == ConstHeaderKey {
			// 获取头数据长度
			messageLength := BytesToInt(buffer[i+ConstHeaderLength+ConstControlLength : i+ConstHeaderLength+ConstControlLength+ConstDataLength])
			
			log.Println("数据长度为:", messageLength)
			log.Println("需要的包长度:", ConstHeaderLength+ConstControlLength+ConstDataLength+messageLength)
			if length < i+ConstHeaderLength+ConstControlLength+ConstDataLength+messageLength {
				break
			}
			data := buffer[i+ConstHeaderLength+ConstControlLength+ConstDataLength : i+ConstHeaderLength+ConstControlLength+ConstDataLength+messageLength]
			controlID := string(buffer[i+ConstHeaderLength : i+ConstHeaderLength+ConstControlLength])
			key := string(buffer[i+len(ConstHeaderKey) : i+ConstHeaderLength])
			log.Println("控制ID为：", controlID, "key为", key)
			dataPackage := DataPackage{
				Data:      data,
				ControlID: controlID,
				Key:       key,
			}
			Rch <- dataPackage

			i += ConstHeaderLength + ConstControlLength + ConstDataLength + messageLength - 1
		}
	}

	if i == length {
		return make([]byte, 0)
	}
	return buffer[i:]

}