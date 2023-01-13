package main

import (
	"os"
	"os/signal"
	"syscall"
	"wshttp/gateway"
)

func main() {
	go gateway.InitWsServer()
	c := make(chan os.Signal, 1)                      // 创建一个接收信号的channel
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM) // 监听系统信号
	<-c
}
