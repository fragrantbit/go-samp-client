package main

import (
	"bundle/bitstream"
	"encoding/binary"
	"log"
)

const NetGame = 4057
const MAX_MTU_SIZE = 587

const (
	IDConnectionRequest         = 0xB
	IDAuthKey                   = 0xC
	IDOpenConnectionRequest     = 0x18
	IDOpenConnectionReply       = 0x19
	IDConnectionCookie          = 0x1A
	IDConnectionRequestAccepted = 0x22
	IDBanned                    = 0x24
	IDRpc                       = 0x14
	IDNewIncommingConnection    = 0x1E
)

func (peer *Peer) Send(data []byte, length int) {
	payload := Encrypt(data, length, peer.port, 0)

	if _, err := peer.handle.Write([]byte(payload)); err != nil {
		log.Println("sendto failed", err)
	}
}

func (peer *Peer) RequestConnectionCookie() {

	c := make([]byte, 5)
	var sum uint32 = 0x6969

	c[0] = 24
	c[1] = 0 ^ byte(sum)

	peer.sendData <- c[:2]
}

func (peer *Peer) SendConnectionCookie(data []byte) {
	var cookie uint32 = 0x0000
	var c []byte
	c = make([]byte, 5)
	a := []uint32{uint32(data[1]), uint32(data[2])}

	cookie = uint32((a[1] << 8) | a[0])

	data[0] = 24
	binary.LittleEndian.PutUint16(data[1:3], uint16(cookie^0x6969))

	copy(c[:], data[:])

	peer.sendData <- c[:4]
}

func (peer *Peer) SendConnectionRequest() {

	bs := bitstream.EmptyBitStream()
	bs.WriteByte(11)

	peer.chanBS <- *bs
}

func (peer *Peer) Start() {
	peer.RequestConnectionCookie()
}

func (peer *Peer) SendNIC(merged []byte) {

	outBS := bitstream.EmptyBitStream()

	outBS.WriteByte(IDNewIncommingConnection)
	outBS.Write(merged, (4*8) + (2*8), true)

	peer.chanBS <- *outBS
}

func (peer *Peer) OnConnectionRequestAccepted(data []byte, length int) {

	readBS := bitstream.NewBitStream(data, len(data))

	recvExternalIDBytes, _ := readBS.ReadUint32()
	recvPortBytes, port := readBS.ReadUint16()

	_, playerID := readBS.ReadUint16()

	log.Println("playerID:", playerID, "port:", port)

	merged := MergeBytes(recvExternalIDBytes, recvPortBytes, (4*8+2*8)>>3)
	peer.SendNIC(merged)

	uiSvrChallenge, _ := readBS.ReadBits(4*8, true)
	d3 := binary.LittleEndian.Uint32(uiSvrChallenge)

	peer.Join(d3)
}

func GenerateBitStream(peer *Peer, output *bitstream.BitStream) {

	bitsUsed := output.NumberOfBitsUsed
	saveData := make([]byte, bitsUsed)
	copy(saveData, output.Data[:])

	output.NumberOfBitsUsed = 0

	output.WriteBool0()

	output.WriteUint16(uint16(peer.packetNumber), false)
	output.Write([]byte{8}, 4, true)
	output.WriteBool0()
	output.WriteUint16(uint16(bitsUsed), true)
	output.WriteAlignedBytes(saveData, bitsUsed>>3)

	peer.packetNumber++
}

func (peer *Peer) ProcessPacket(dataBlock *DataBlock) {

	packetID := dataBlock.packetID
	data := dataBlock.data

	switch packetID {

	case IDOpenConnectionReply:
		if !peer.authDone {
			peer.SendConnectionRequest()
		}

	case IDConnectionCookie:

		if !peer.authDone {
			cookie := []byte{}
			cookie = append(cookie, packetID)
			cookie = append(cookie, data...)

			peer.SendConnectionCookie(cookie)
		}

	case IDConnectionRequestAccepted:

		if peer.authDone {
			return
		}
		peer.authKeySent <- true
		peer.OnConnectionRequestAccepted(data, len(data))

	case IDBanned:

		/*if !peer.authDone && peer.isAuthKeyPending {
			peer.authKeySent <- false
		}*/
		//log.Println("banned") // works
		//return

	case IDAuthKey:
		if !peer.authDone {
			peer.SendAuthKey(data)
		}

	case IDRpc:
		peer.HandleRPC(data, len(data))

	case 227:
		return
	case 0:
		return
	default:

	}
}

func (peer *Peer) Join(ui uint32) {

	log.Println("Connected. Joining the game as", peer.Client.Name)

	outcomingBS := bitstream.EmptyBitStream()

	var version uint32 = uint32(NetGame)

	challengeResponse := ui ^ version

	authBsKey := []byte("12616EE8D60CF543732647C8F08F2997E8D084D5401")
	authBsKeyLen := len(authBsKey)

	botNameLen := len(peer.Client.Name)

	outcomingBS.WriteUint32(uint32(version), false)
	outcomingBS.WriteByte(1)

	outcomingBS.WriteByte(byte(botNameLen))
	outcomingBS.Write([]byte(peer.Client.Name), botNameLen*8, true)

	outcomingBS.WriteUint32(challengeResponse, false)

	outcomingBS.WriteByte(byte(authBsKeyLen))
	outcomingBS.WriteAnArray(authBsKey, authBsKeyLen)

	peer.SendRPC(25, *outcomingBS)
}

func (peer *Peer) SendAuthKey(data []byte) {

	newBS := bitstream.NewBitStream(data, len(data))

	authLen, _ := newBS.ReadByte()

	auth := newBS.ReadAnArray(int(authLen))
	authStr := string(auth)
	authKey, found := FindAuthKey(authStr, int(authLen))

	if !found {
		log.Println("Failed to find authkey.")
		return
	}

	bs := bitstream.EmptyBitStream()

	bs.WriteByte(IDAuthKey)
	bs.WriteByte(byte(len(authKey)))
	bs.WriteAnArray([]byte(authKey), len(authKey))

	peer.chanBS <- *bs
	log.Println("Sending auth key...")

	do := func() { peer.sendData <- bs.Data[:(bs.NumberOfBitsUsed>>3)+1] }

	NewTask(do, &peer.authKeySent, &peer.authDone, 5, true)

}
