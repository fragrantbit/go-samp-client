package main

func (peer *Peer) InitializeHeart() {
	peer.data = make(chan []byte)
	go peer.Reader()
	peer.Processor()
}

func (peer *Peer) Processor() {

	for {
		select {
		case data := <-peer.data: {		
			if block := CreateBlock(data, len(data)); block != nil {
				peer.ProcessPacket(block)
			}
		}
		
		default:

		}
	}
}

func (peer *Peer) Reader() {
	buf := make([]byte, 586)
	
	for {
		n, _ := peer.handle.Read(buf)		
		peer.data <- buf[:n]
	}
}