package main

func (peer *Peer) InitializeHeart() {
	peer.recvData = make(chan []byte, 5)
	peer.sendData = make(chan []byte, 5)
	go peer.Reader()
	peer.Processor()
}

func (peer *Peer) Processor() {

	go func() {
		for {
			select {
			case data := <-peer.recvData: {		
				if block := CreateBlock(data, len(data)); block != nil {
					peer.ProcessPacket(block)
				}
			}
			default:

			}
		}
	}()

	
	for {
		select {
		case data := <-peer.sendData: {
			peer.Send(data, len(data))
		}
		case bs := <-peer.chanBS: {
			GenerateBitStream(peer, &bs)
			peer.sendData <- bs.Data
		}
		default:
			
		}
	}
	
}

func (peer *Peer) Reader() {
	buf := make([]byte, MAX_MTU_SIZE)
	for {
		n, _ := peer.handle.Read(buf)			
		peer.recvData <- buf[:n]
	}
}