package proxy 

import (
	"fmt"
)

type ProxyConfig struct{
	Listen   string `toml:"listen"`    // Proxy 服务的地址 
}

var Config *ProxyConfig

func InitConfig() (err error){

	if Config == nil {

		// 加载配置信息
		fileName = "proxy.conf"
		if err01 := util.DecodeSection(fileName, sectionName, Config); err != nil {
			err = fmt.Errorf("Load config file failed, error:%s", err.Error())			
			return 
		}

		if Config.Listen == "" {
			err = fmt.Errorf("侦听地址为空，请在配置文件proxy.conf配置listen值")			
			return 
		}

	}
	
	return nil 
}	