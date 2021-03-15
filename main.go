package main

import (
	"log"
)


func main() {
	peer := InitPeer("0.0.0.0", 7777)
	Init()
	log.Println("Started")
	go peer.Start()
	peer.InitializeHeart()
}

