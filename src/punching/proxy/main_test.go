package proxy_test

import (
	"net"
	"punching/client"

	"io"
	//	"io/ioutil"
	"net/http"
	"os"
	"punching/logger"
	"punching/proxy"
	"punching/server"
	"testing"
	"time"
)

// var WG sync.WaitGroup

const (
	RENDER_FILE_PATH = "/Users/chenboxing/nat/src/punching/index.html"
)

func runHttpWeb(addr string) {

	//第一个参数为客户端发起http请求时的接口名，第二个参数是一个func，负责处理这个请求。
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {

		f, _ := os.Open(RENDER_FILE_PATH)
		defer f.Close()
		//读取页面内容
		io.Copy(w, f)
	})

	//服务器要监听的主机地址和端口号
	err := http.ListenAndServe(addr, nil)

	if err != nil {
		logger.Errorf("ListenAndServe error: ", err.Error())
	} else {
		logger.Infof("开启服务在:%s", addr)
	}

}

func TestNat(t *testing.T) {

	// 开启Proxy服务
	go func() {
		proxy.Main()
	}()

	for {
		time.Sleep(2 * time.Second)
		conn, err := net.Dial("tcp", proxy.Config.Listen)
		if err != nil {
			time.Sleep(2 * time.Second)
		} else {
			conn.Close()
			break
		}
	}

	// Server连接
	go func() {
		server.Main()
	}()

	server.InitConfig()
	pairName := server.Config.Key

	logger.Infof("Pairname is :%s, %+v", pairName, server.Config)
	// Check if the P2P server is available
	for {
		logger.Info(len(proxy.OnlineServerList))
		if _, ok := proxy.OnlineServerList[pairName]; !ok {
			time.Sleep(2 * time.Second)
			continue
		}
		break
	}

	logger.Info("准备开启客户端")

	go func() {
		client.Main()
	}()

	//client.InitConfig()
	//for {
	//	time.Sleep(2 * time.Second)
	//	logger.Infof("准备连接:%s", client.Config.Listen)
	//	conn, err := net.Dial("tcp", client.Config.Listen)
	//	if err != nil {
	//		time.Sleep(2 * time.Second)
	//	} else {
	//		conn.Close()
	//		break
	//	}
	//}
	//
	//// 启用Web后台服务
	//server.Config.Dial = ":7779"
	//
	//logger.Infof("开启后台服务:%s",server.Config.Dial)
	////开启服务
	//go runHttpWeb(server.Config.Dial)
	//
	//url := "http://" + client.Config.Listen
	//logger.Infof("获取网页内容:")
	//
	//if resp, err := http.Get(url); err != nil {
	//	t.Errorf("读取配置端口出错,%s", client.Config.Listen)
	//} else {
	//	defer resp.Body.Close()
	//	arrContent, _ := ioutil.ReadAll(resp.Body)
	//
	//	if all, err := ioutil.ReadFile(RENDER_FILE_PATH); err != nil {
	//		t.Errorf("读取文件出现错误, %s", err)
	//	} else {
	//		if len(all) != len(arrContent) {
	//			t.Errorf("文件大小不一致:%d,%d", len(all), len(arrContent))
	//		}
	//	}
	//
	//}

	select {}
}
