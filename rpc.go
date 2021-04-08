package main

import (
	"bundle/bitstream"
	"fmt"
)

func (peer *Peer) SendRPC(rpcID int, bs bitstream.BitStream) {

	bitlen := bs.NumberOfBitsUsed
	input := bitstream.EmptyBitStream()

	input.WriteByte(IDRpc)
	input.WriteByte(byte(rpcID))
	input.WriteUint32(uint32(bitlen), true)

	input.Write(bs.Data, bitlen, false)

	peer.chanBS <- *input
}

func (peer *Peer) HandleRPC(data []byte, length int) {
	
	bs := bitstream.NewBitStream(data, length)

	test := bs.ReadAnArray(1)
	fmt.Println(test)
}


