package main

import (
	"fmt"
	"net"
	"sync"
)

type Server struct {
	Ip   string
	Port int
	// 在线用户列表
	OnlineMap map[string]*User
	mapLock   sync.RWMutex
	// 消息广播的channel
	Message chan string
}

// 创建一个server的接口

func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}
	return server
}

// 监听Message广播消息channel的goroutine，一旦有消息就发送给全部的在线user
func (server *Server) ListenMessager() {
	for {
		msg := <-server.Message
		// 将msg发送给全部的在线user
		server.mapLock.Lock()
		for _, cli := range server.OnlineMap {
			cli.C <- msg
		}
		server.mapLock.Unlock()
	}
}

// 广播消息
func (server *Server) BroadCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg
	server.Message <- sendMsg
}

func (server *Server) Handler(conn net.Conn) {
	// 当前连接的业务
	// 用户上线，将用户加入到OnlineMap中
	server.mapLock.Lock()

	user := NewUser(conn)
	server.OnlineMap[user.Name] = user
	server.mapLock.Unlock()

	// 广播当前用户上线消息
	server.BroadCast(user, "on line")

}

// 启动服务器的接口
func (server *Server) Start() {
	// socket listen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", server.Ip, server.Port))
	if err != nil {
		fmt.Println("net.Listen err:", err)
		return
	}
	// close listen socket
	defer listener.Close()

	// 启动监听Message的goroutine
	go server.ListenMessager()

	for {
		// accept
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listener.Accept err:", err)
			continue
		}
		// do handler
		go server.Handler(conn)
	}

}
