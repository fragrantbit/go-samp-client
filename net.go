package main

import (
	"net"
	"fmt"
	"strconv"
)
type Peer struct {
	handle 					net.Conn
	port   					uint32
	addr   					string
	data   					chan []byte

	packetNumber 			int
	sockfd 					int
	authDone 				bool
	authKeySent				chan bool

	//dst  				syscall.SockaddrInet4
	
}

func InitPeer(addr string, port uint32) *Peer {
	conn, err := net.Dial("udp4", addr + ":" + strconv.Itoa(int(port)))
	if err != nil {
		fmt.Println("auth: net.Dial failed")
		return &Peer{}
	}

	return &Peer{
		port: port, 
		handle: conn,
		addr: addr,
		packetNumber: 0,
		authDone: false,
		authKeySent: make(chan bool, 4),
	}
}
