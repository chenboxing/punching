.PHONY : default server client proxy all_windows all_darwin windows arm darwin deps fmt  clean all
export GOPATH:=$(shell pwd)

PREFIX=''
default: all

GOOS=
GOARCH=
GOARM=

fmt:
	go fmt punching/...

deps:
	go get  -d -v punching/...

server: deps
	go install punching/main/server

client: deps
	go install punching/main/client

proxy: deps
	go install  punching/main/proxy


server_linux:
	GOOS=linux GOARCH=amd64 go install punching/main/server
client_linux:
	GOOS=linux GOARCH=amd64 go install punching/main/client
proxy_linux:
	GOOS=linux GOARCH=amd64 go install punching/main/proxy

server_windows:
	GOOS=windows GOARCH=amd64 go install punching/main/server
client_windows:
	GOOS=windows GOARCH=amd64 go install punching/main/client
proxy_windows:
	GOOS=windows GOARCH=amd64 go install punching/main/proxy

server_darwin:
	GOOS=darwin GOARCH=amd64 go install punching/main/server
client_darwin:
	GOOS=darwin GOARCH=amd64 go install punching/main/client
proxy_darwin:
	GOOS=darwin GOARCH=amd64 go install punching/main/proxy

server_arm:
	GOOS=linux GOARCH=arm  GOARM=5  go install punching/main/server
client_arm:
	GOOS=linux GOARCH=arm  GOARM=5  go install punching/main/client
proxy_arm:
	GOOS=linux GOARCH=arm  GOARM=5 go install punching/main/proxy


all_darwin: fmt client_darwin server_darwin proxy_darwin
all_linux: fmt client_linux server_linux proxy_linux
all_windows: fmt client_windows server_windows proxy_windows
all_arm: fmt client_arm server_arm proxy_arm
all_platform: all_darwin all_linux all_windows all_arm
all: fmt client server proxy

clean:
	go clean -i -r punching/...
