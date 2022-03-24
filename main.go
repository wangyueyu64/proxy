package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"net/url"
)

func main() {

	if err := LogInit(); err != nil {
		fmt.Printf("init log failed err :%v\n", err)
	}
	l, err := net.Listen("tcp", ":8081")
	if err != nil {
		log.Panic(err)
	}
	for {
		client, err := l.Accept()
		if err != nil {
			log.Panic(err)
		}
		go handleClientRequest(client)
	}
}func handleClientRequest(client net.Conn) {

	var b [1024]byte
	var method, host, address string

	if client == nil {
		return
	}
	defer client.Close()

	n, err := client.Read(b[:])
	if err != nil {
		log.Println(err)
		return
	}

	fmt.Sscanf(string(b[:bytes.IndexByte(b[:], '\n')]), "%s%s", &method, &host)
	hostPortURL, err := url.Parse(host)
	if err != nil {
		log.Println(err)
		return
	}
	//http访问
	address = hostPortURL.Host
	//获得了请求的host和port，就开始拨号吧
	server, err := net.Dial("tcp", address)
	if err != nil {
		log.Println(err)
		return
	}
	if method == "CONNECT" {
		fmt.Fprint(client, "HTTP/1.1 200 Connection established\r\n")
	} else {
		server.Write(b[:n])
	}    //进行转发
	go io.Copy(server, client)
	io.Copy(client, server)
}
