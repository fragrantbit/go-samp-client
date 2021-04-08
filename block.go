package main

import (
	"bundle/bitstream"
)

type DataBlock struct {
	packetID byte
	data     []byte
}

func UnpackPacket(data []byte, length int) ([]byte, bool) {

	readBS := bitstream.NewBitStream(data, length)

	var hasAcks bool
	t := readBS.ReadBool(&hasAcks)

	if t && hasAcks {
		if readBS.DeserializeBitStream() != true {
			return nil, false
		}
	} else {
		return nil, false
	}

	_, success := readBS.ReadBits(2*8, true)
	if success != true {
		return nil, false
	}

	_, success = readBS.ReadBits(4, true)
	if success != true {
		return nil, false
	}

	var iss bool
	success = readBS.ReadBool(&iss)
	if success != true {
		return nil, false
	}

	dataBitLen, success := readBS.ReadCompressed(16, true)
	if success != true {
		return nil, false
	}
	if len(dataBitLen) == 0 {
		return nil, false
	}

	content := readBS.ReadAlignedBytes(bitstream.BitsToBytes(int(dataBitLen[0])))
	if content == nil {
		return nil, false
	}

	return content, true
}

func CreateBlock(data []byte, length int) *DataBlock {

	if content, success := UnpackPacket(data, length); !success {

		dataBlock := new(DataBlock)

		dataBlock.packetID = data[0]
		dataBlock.data = append(dataBlock.data, data[1:]...)

		return dataBlock

	} else {
		dataBlock := new(DataBlock)

		dataBlock.packetID = content[0]
		dataBlock.data = append(dataBlock.data, content[1:]...)

		return dataBlock
	}
}

func MergeBytes(source []byte, source2 []byte, size int) []byte {

	result := make([]byte, size)

	copy(result, source)
	copy(result[len(result):], source2)

	return result
}
