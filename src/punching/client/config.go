package client

import (
	"fmt"
	"os"
	"punching/util"
)

/*
[server]
proxy  = ""
dial   = ""
key    = ""

[ThirdProxy]
address  = "proxy.move8.cn:7777"
*/

type ClientConfig struct {
	Proxy  string `toml:"proxy"`  // Proxy 服务的地址
	Listen string `toml:"listen"` // 服务端提供的服务地址
	Key    string `toml:"key"`    // 客户端和服务端的匹配码
}

type ThirdProxyConfig struct {
	Address string `toml:"address"` // Proxy 服务的地址
}

var Config ClientConfig
var ThirdConfig ThirdProxyConfig

func InitConfig() (err error) {

	// 加载配置信息
	fileName := "client.conf"

	if os.Getenv("CLIENT_CONF") != "" {
		fileName = os.Getenv("CLIENT_CONF")
	}

	sectionName1 := "client"
	if err01 := util.DecodeSection(fileName, sectionName1, &Config); err01 != nil {
		err = fmt.Errorf("Load config file failed, error:%s", err01.Error())
		return
	}

	sectionName2 := "ThirdProxy"
	if err02 := util.DecodeSection(fileName, sectionName2, &ThirdConfig); err != nil {
		err = fmt.Errorf("Load config file failed, error:%s", err02.Error())
		return
	}

	if Config.Key == "" {
		err = fmt.Errorf("匹配码不能为空，请在client.conf配置匹配码")
		return
	}

	if Config.Proxy == "" && ThirdConfig.Address == "" {
		err = fmt.Errorf("Proxy服务地址和第三方Proxy服务地址不能同时为空")
		return
	}

	return nil
}
