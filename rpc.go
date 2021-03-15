package main

import (
	"bundle/bitstream"
	"fmt"
)

/*
	1: 20 25 220 0 59 33 224 0 0 32 232 236 45 204 140 45 140 218 234 162 232 37 70 38 6 102 166 6 8 38 232 102 72 104 167 6 134 71 6 40 72 134 40 166 230 38 200 166 70 72 102 104 200 72 38 71 40 168 198 134 134 232 166 8 192
	1: 20 25 221 0 59 33 224 0 0 32 232 236 45 204 140 45 140 218 234 162 232 37 102 38 8 70 40 38 104 102 8 198 72 167 38 200 200 166 40 72 104 198 166 6 166 70 40 70 200 72 103 39 7 38 166 38 40 40 136 198 38 71 39 8 104 128

	2: 1 128 64 22 4 20 25 221 0 59 33 224 0 0 32 232 236 45 204 140 45 140 218 234 162 232 37 102 38 70 198 38 200 168 167 8 134 198 8 104 198 166 134 102 230 102 70 198 134 232 103 8 198 7 8 198 71 39 38 232 167 8 134 7 6 136 134 166 134 6 32
	2: 1 128 64 22 4 20 25 221 0 59 33 224 0 0 32 232 236 45 204 140 45 140 218 234 162 232 37 102 38 70 198 38 200 168 167 8 134 198 8 104 198 166 134 102 230 102 70 198 134 232 103 8 198 7 8 198 71 39 38 232 167 8 134 7 6 136 134 166 134 6 32
*/

func (peer *Peer) SendRPC(rpcID int, bs bitstream.BitStream) {

	bitlen := bs.NumberOfBitsUsed
	input := bitstream.EmptyBitStream()

	input.WriteByte(IDRpc)
	input.WriteByte(byte(rpcID))
	input.WriteUint32(uint32(bitlen), true)

	input.Write(bs.Data, bitlen, false)

	GenerateBitStream(peer, input)

	peer.Send(input.Data, len(input.Data))
}

func (peer *Peer) HandleRPC(data []byte, length int) {
	incomingBS := bitstream.NewBitStream(data, length)
	uniqueID := 0
	_ = uniqueID
	uniqueIDBytes := incomingBS.ReadAnArray(1)
	fmt.Println(uniqueIDBytes)
	if uniqueIDBytes[0] == 128 {
		requestOutcome, _ := incomingBS.ReadBits(8, true)
		fmt.Println(requestOutcome)
	}
}
