package main

import (
	"bytes"
	"fmt"
	"io"
	"net"
)

func main() {

	if err := LogInit(); err != nil {
		fmt.Printf("init log failed err :%v\n", err)
	}
	l, err := net.Listen("tcp", ":8080")
	if err != nil {
		logger.Errorf("Listen err :%v", err)
	}
	for {
		client, err := l.Accept()
		if err != nil {
			logger.Errorf("Accept err :%v", err)
		}
		go handleClientRequest(client)
	}
}
func handleClientRequest(client net.Conn) {

	var b [1024]byte
	var method, host string
	targetAddress := "localhost:8081"

	if client == nil {
		return
	}
	defer client.Close()

	n, err := client.Read(b[:])
	if err != nil {
		logger.Errorf("Read err :%v", err)
		return
	}

	_, err = fmt.Sscanf(string(b[:bytes.IndexByte(b[:], '\n')]), "%s%s", &method, &host)
	if err != nil {
		logger.Errorf("Sscanf err :%v", err)
		return
	}

	//获得了请求的host和port，就开始拨号吧
	server, err := net.Dial("tcp", targetAddress)
	if err != nil {
		logger.Errorf("Dial err :%v", err)
		return
	}
	if method == "CONNECT" {
		fmt.Fprint(client, "HTTP/1.1 200 Connection established\r\n")
	} else {
		server.Write(b[:n])
	} //进行转发
	go io.Copy(server, client)
	io.Copy(client, server)
}
