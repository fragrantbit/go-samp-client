package main

import (
	"log"
)


func main() {
	peer := InitPeer("127.0.0.1", 7777, "Nickname")
	Init()
	log.Println("Started")
	go peer.Start()
	peer.InitializeHeart()
}

