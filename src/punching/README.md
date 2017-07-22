### 使用场景 

  


### 项目介绍
   
   本项目完全是点对点的访问，无须经由服务器转发，代理转发端只是为了获取当自的Nat网络IP和端口，为了Nat穿透,在获取各自对方的IP和端口后，就不需要服务端的干预。
   
### 如何使用
  1. 迁出源码git clone https://github.com/chenboxing/punching.git 
  2. 进入项目目录 src/punching/
  3. 编译源码 
    make (all|server|client|proxy)  (windows|darwin|linux|arm)
    make all windows           #编译windows平台下所有
    make server|client linux   #编译Windows平台Server端和Client端二进制文件
    二进制文件编译在 punching/bin/里

    你也可以直接下载已经编译好的文件：
       

  4. 把proxy(代理转发端) 部署到公网计算机上，配置config.conf配置节[proxy],设置侦听端口,默认7777，如果没有公网计算机，可以使用本站设置好的域名 nat.move8.cn, 在config.conf配置节[ThridProxy]设置好相关信息，比如
    [ThirdProxy]
    address = nat.move8.cn 
    email   = xxxx@xxxxx.com   
    password = xxxxxxx
   
    先在Nat网络一端，你需要开放访问的服务的计算机上部署server端，配置config.conf配置节[server],在listen项里设置你要开放的应用服务，如 192.168.1.45:80, proxy添写你的代理转发端公网地址和端口，比如 xxx.f3322.net:7777, 如果你需要使用本站提供的代理转发，此项请为空,参考上面填写节[ThirdProxy]信息。

    配置好，启动Server端:
    nat_server.exe 

    配置在Client端
    在Nat网络的另一端，部署client端，先配置config.conf,在节[client],需要先设置好要侦听的端口信息，比如listen = :8585, proxy添写你的代理转发端公网地址和端口，比如 xxx.f3322.net:7777, 如果你需要使用本站提供的代理转发，此项请为空,参考上面填写节[ThirdProxy]信息。

    配置好，启动Server端:
    nat_client.exe 
    
### 局限
  只能P2P通讯(Client端,Server端),无法多人访问开启的服务
  