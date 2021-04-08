package main

import (
	"net"
	"fmt"
	"strconv"
	"bundle/bitstream"
)
type Peer struct {
	handle 					net.Conn
	port   					uint32
	addr   					string

	recvData   				chan []byte
	sendData				chan []byte
	chanBS					chan bitstream.BitStream
	authKeySent				chan bool

	packetNumber 			int
	authDone 				bool


	Client *Client
}

type Client struct {
	Name string
} 

func InitPeer(addr string, port uint32, name string) *Peer {
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
		chanBS: make(chan bitstream.BitStream),
		Client: &Client{
			Name: name,
		},
	}
}
