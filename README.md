### 使用场景 
   
有时，我们需要在家里访问公司的电脑，但如果任一方路由器分配的IP是运营商的内网IP,而不是公网IP,双方就无法直接通信。为了让双方电脑可以通讯，我们可以借助第三方提供的tunnel服务，利用服务器转发进行通讯，比较出名的是花生壳的内网穿透服务。本项目实现是的另外的解决思路，基于P2P NAT穿透，使位于NAT网络后的计算机直接通讯，无需借助第三方服务器进行数据转发，简单，高效和安全。

### 项目介绍
   
    ![image](https://github.com/chenboxing/punching/raw/bak/screenshots/architecture.png)
   
   
   如上图所示，在公网上部署一台跨网服务器，该服务器运行解析端(proxy), P2P客户端和服务端启动时向跨网解析端提交TCP连接请求，以便公网解析端根据请求，记录各自自连接时NAT地址，并告知对方的NAT地址。P2P客户端和服务端尝试同时连接，进行NAT穿透。在穿透成功后，P2P终端可以脱离跨网解析端独立进行TCP数据通讯，无需第三方数据转发。
   
### 如何使用

  1. 迁出源码 git clone https://github.com/chenboxing/punching.git 
  2. 进入项目目录 src/punching/
  3. 编译源码 
    make all   # 编译所在平台的所有端
    make_windows # 编译windows平台的所有端
    make_linux   # 编译linux平台的所有端
    make_darwin  # 编译mac os 平台的所有端
    make_arm     # 编译arm嵌入式平台的所有端，arm版本基于5
    
    编译后二进制文件放在 punching/bin/目录下
       

  4. 跨网解析端和P2P端配置和使用
     
     4.1 跨网解析端部署(如没有公网服务器，此步可跳过)：
        
     把proxy(代理转发端)和配置文件proxy.conf 部署到公网计算机上，配置proxy.conf配置节[proxy],设置侦听端口,默认7777
     运行解析端
        ./proxy
           
     4.2 配置P2P服务端和客户端
        
    [ThirdProxy]
    address = nat.move8.cn 
    email   = xxxx@xxxxx.com   
    password = xxxxxxx
   
    先在Nat网络一端，你需要开放访问服务的计算机上部署server端，配置config.conf配置节[server],在listen项里设置你要开放的应用服务，如 192.168.1.45:80, proxy添写你的代理转发端公网地址和端口，比如 xxx.f3322.net:7777, 如果你需要使用本站提供的代理转发，此项请为空,参考上面填写节[ThirdProxy]信息。

    配置好，启动Server端:
    nat_server.exe 

    配置在Client端
    在Nat网络的另一端，部署client端，先配置config.conf,在节[client],需要先设置好要侦听的端口信息，比如listen = :8585, proxy添写你的代理转发端公网地址和端口，比如 xxx.f3322.net:7777, 如果你需要使用本站提供的代理转发，此项请为空,参考上面填写节[ThirdProxy]信息。

    配置好，启动Server端:
    nat_client.exe 
    
### 当前局限
  1. 只能提供一对一的P2P通讯(Client端,Server端),无法多人访问开启的P2P服务端服务.如果你想实现多人访问开启的服务，你可以使用tunnel(隧道)工具，但前提是需要架设一台线上服务器，部署tunnel服务端，在要开放服务的计算机上部署tunnel客户端。
  2. 现在，项目只实现了TCP P2P穿透方案,如果后台服务需要UDP协议通讯 ，无法工作，比如vnc服务就无法访问，需要实现UDP和TCP穿透才可以工作。windows xp sp2下的平台也不支持P2P连接的TCP同时连接特性，在xp sp2平台上无法工作。
  
  
### 依赖的第三方包  
   github.com/cihub/seelog： 日志记录增加包
   github.com/BurntSushi/toml： toml配置文件处理包
