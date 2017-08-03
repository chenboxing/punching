package proxy

import (
	"fmt"
	"os"
	"punching/util"
)

type ProxyConfig struct {
	Listen string `toml:"listen"` // Proxy 服务的地址
}

var Config ProxyConfig

func InitConfig() (err error) {

	// 加载配置信息
	// fileName := "/Users/chenboxing/nat/src/punching/src/punching/proxy.conf"
	fileName := "proxy.conf"
	if os.Getenv("PROXY_CONF") != "" {
		fileName = os.Getenv("PROXY_CONF")
	}
	sectionName := "proxy"
	if err01 := util.DecodeSection(fileName, sectionName, &Config); err01 != nil {
		err = fmt.Errorf("Load config file failed, error:%s", err01.Error())
		return
	}

	if Config.Listen == "" {
		err = fmt.Errorf("侦听地址为空，请在配置文件proxy.conf配置listen值")
		return
	}

	return nil
}
