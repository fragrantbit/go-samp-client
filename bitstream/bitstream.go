package bitstream

import (
	"encoding/binary"
)

/*
	Brief implementation of RakNet BitStream.

	This is written for exchanging required data for connection to SA:MP server and subsequent data exchange.
	
	Based on the original RakNet source code.
*/

type BitStream struct {
	Data []byte
	NumberOfBitsUsed int
	NumberOfBitsAllocated int
	ReadOffset int

	Data2 []int16
}

func NewBitStream(data []byte, length int) *BitStream {
	return &BitStream{
		Data: data,
		NumberOfBitsUsed: length << 3,
		ReadOffset: 0,
		NumberOfBitsAllocated: length << 3,
	}
}

func EmptyBitStream() *BitStream {
	return &BitStream{
		NumberOfBitsAllocated: 256 * 8,
		NumberOfBitsUsed: 0,
		Data: make([]byte, 587),
	}
}

func (bs *BitStream) WriteAnArray(data []byte, bytesToWrite int) {
	if (bs.NumberOfBitsUsed & 7) == 0 {
		copy(bs.Data[BitsToBytes(bs.NumberOfBitsUsed):], data)
		bs.NumberOfBitsUsed += BytesToBits(bytesToWrite)	
	} else {
		bs.Write(data, bytesToWrite * 8, true)
	}
}

func (bs *BitStream) WriteUint32(data uint32, compressed bool) {
	bytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(bytes, data)

	if compressed {
		bs.WriteCompressed(bytes, 4 * 8, true)
	} else {
		bs.Write(bytes, 4 * 8, true)
	}
}

func (bs *BitStream) WriteUint16(data uint16, compressed bool) {
	bytes := make([]byte, 2)

	binary.LittleEndian.PutUint16(bytes, data)

	if compressed {
		bs.WriteCompressed(bytes, 2 * 8, true)
	} else {
		bs.Write(bytes, 2 * 8, true)
	}
}

func (bs *BitStream) AddBitsAndReallocate(num int) {
	
}

func (bs *BitStream) Write(data []byte, bitsToWrite int, rightAlignedBits bool) {

	offset := 0
	
	numberOfBitsUsedMod8 := bs.NumberOfBitsUsed & 7
	for bitsToWrite > 0 {
		dataByte := data[offset]
		if bitsToWrite < 8 && rightAlignedBits {
			dataByte <<= 8 - bitsToWrite
		}
		if numberOfBitsUsedMod8 == 0 {
			bs.Data[bs.NumberOfBitsUsed >> 3] = dataByte
		} else {
			bs.Data[bs.NumberOfBitsUsed >> 3] |= dataByte >> (numberOfBitsUsedMod8)
			if 8 - (numberOfBitsUsedMod8) < 8 && 8 - (numberOfBitsUsedMod8) < bitsToWrite {
				bs.Data[(bs.NumberOfBitsUsed >> 3) + 1] = byte(dataByte << (8 - (numberOfBitsUsedMod8)))
			}
		}
		if bitsToWrite >= 8 {
			bs.NumberOfBitsUsed += 8
		} else {
			bs.NumberOfBitsUsed += bitsToWrite
		}
		bitsToWrite -= 8
		offset++
	}
}

func (bs *BitStream) WriteBool0() {

	numberOfBitsMod8 := bs.NumberOfBitsUsed & 7

	if numberOfBitsMod8 == 0 {
		bs.Data[bs.NumberOfBitsUsed >> 3] = 0
	}

	bs.NumberOfBitsUsed++
}

func (bs *BitStream) WriteBool1() {
	numberOfBitsMod8 := bs.NumberOfBitsUsed & 7

	if numberOfBitsMod8 == 0 {
		bs.Data[bs.NumberOfBitsUsed >> 3] = 0x80
	} else {
		bs.Data[bs.NumberOfBitsUsed >> 3] |= 0x80 >> (numberOfBitsMod8)
	}

	bs.NumberOfBitsUsed++
}

func BitsToBytes(x int) int {
	return (((x)+7)>>3)
}

func BytesToBits(x int) int {
	return (x<<3)
}

func (bs *BitStream) AlignWriteToByteBoundary() {
	if bs.NumberOfBitsUsed != 0 {
		bs.NumberOfBitsUsed += 8 - (((bs.NumberOfBitsUsed - 1) & 7) + 1)
	}
}

func (bs *BitStream) WriteAlignedBytes(data []byte, bytesToWrite int) {
	bs.AlignWriteToByteBoundary()

	bs.WriteAnArray(data, bytesToWrite)
}

func (bs *BitStream) WriteCompressed(data []byte, size int, unsignedData bool) {
	currentByte := (size >> 3) - 1
	var byteMatch byte

	if unsignedData {
		byteMatch = 0
	} else {
		byteMatch = 0xFF
	}

	for currentByte > 0 {
		if data[currentByte] == byteMatch {
			bs.WriteBool1()
		} else {
			bs.WriteBool0()

			bs.Write(data, (currentByte + 1) << 3, true)

			return
		}

		currentByte--
	}

	if unsignedData && data[currentByte] & 0xF0 == 0x00 || unsignedData == false && data[currentByte] & 0xF0 == 0xF0 {
		bs.WriteBool1()
		b := []byte{data[currentByte]}
		bs.Write(b, 4, true)
	} else {
		bs.WriteBool0()
		b := []byte{data[currentByte]}
		bs.Write(b, 8, true)
	}
}

func (bs *BitStream) DeserializeBitStream() bool {

	count, ok := bs.ReadCompressed(2*8, true)
	var maxEqualToMin bool
	var min byte
	var max byte
	var maxres []byte
	var minres []byte
	if !ok {
		return false
	}
	for i := 0; i < int(count[0]); i++ {
		bs.ReadBool(&maxEqualToMin)
		minres, ok = bs.ReadBits(2*8, true)
		if !ok {
			return false
		}
		min = minres[0]
		if maxEqualToMin == false {
			maxres, ok = bs.ReadBits(2*8, true)
			if !ok {
				return false
			}
			max = maxres[0]
			if max < min {
				return false
			}

		} else {
			max = min
		}
	}

	return true
}

func (bs *BitStream) ReadBool(v *bool) bool {
	if bs.ReadOffset + 1 > bs.NumberOfBitsUsed {
		return false
	}
	if bs.Data[bs.ReadOffset >> 3] & (0x80 >> (bs.ReadOffset % 8)) != 0 {
		*v = true
	} else {
		*v = false
	}
	bs.ReadOffset++
	return true
}

func (bs *BitStream) ReadCompressed(size int, unsignedData bool) ([]byte, bool) {
	currentByte := (size >> 3) - 1

	output := make([]byte, size)

	byteMatch := 0
	halfByteMatch := 0

	if unsignedData {
		byteMatch = 0
		halfByteMatch = 0
	} else {
		byteMatch = 0xFF
		halfByteMatch = 0xF0
	}

	for currentByte > 0 {
		var b bool

		bs.ReadBool(&b)
		if b {
			output[currentByte] = byte(byteMatch)
			currentByte--
		} else {
			
			output2, ok := bs.ReadBits((currentByte + 1) << 3, true)
			if !ok {
				return nil, false
			}
			return output2, true
		}
	}
	
	var b bool = true
	var output2 []byte
	bs.ReadBool(&b)
	if !b {
		output2, _ = bs.ReadBits(8, true) 
	} else {
		output2, _ = bs.ReadBits(4, true)
		output2[currentByte] |= byte(halfByteMatch)
	}
	return output2, true
}

func (bs *BitStream) AlignReadToByteBoundary() {
	if bs.ReadOffset > 0 {
		bs.ReadOffset += 8 - (((bs.ReadOffset - 1) & 7) + 1)
	}
}

func (bs *BitStream) ReadAnArray(bytes int) []byte {
	output := make([]byte, bytes)
	if bs.ReadOffset & 7 == 0 {
		output = bs.Data[bs.ReadOffset >> 3:(bs.ReadOffset >> 3) + bytes]
		bs.ReadOffset += bytes << 3
		return output
	} else {
		output, _ = bs.ReadBits(bytes * 8, true)
		return output
	}
}

func (bs *BitStream) ReadAlignedBytes(bytesToRead int) []byte {

	if bytesToRead <= 0 {
		return nil
	}
	bs.AlignReadToByteBoundary()
	if bs.ReadOffset + (bytesToRead << 3) > bs.NumberOfBitsUsed {
		return nil
	}

	output := make([]byte, bytesToRead)

	copy(output, bs.Data[bs.ReadOffset >> 3:])
	bs.ReadOffset += bytesToRead << 3

	return output
}

func (bs *BitStream) ReadBits(bitsToRead int, alignBitsToRight bool) ([]byte, bool) {
	if bitsToRead <= 0 { 
	  return nil, false
	}
	if bs.ReadOffset + bitsToRead > bs.NumberOfBitsUsed {
		return nil, false
	}
	offset := 0

	readOffsetMod8 := bs.ReadOffset & 7
	output := make([]byte, BitsToBytes(bitsToRead))

	for bitsToRead > 0 {
		output[offset] |= bs.Data[bs.ReadOffset >> 3] << readOffsetMod8

		if readOffsetMod8 > 0 && bitsToRead > 8 - (readOffsetMod8) {
			output[offset] |= bs.Data[(bs.ReadOffset >> 3) + 1] >> (8 - (readOffsetMod8))
		}

		bitsToRead -= 8

		if bitsToRead < 0 {
			if alignBitsToRight {
				output[offset] >>= -bitsToRead
			}

			bs.ReadOffset += 8 + bitsToRead
		} else {
			bs.ReadOffset += 8
		}

		offset++
	}

	return output, true
}

func (bs *BitStream) ReadByte() (byte, bool) {
	result, ok := bs.ReadBits(8, true)

	if !ok {
		return 0, false
	}

	return result[0], false
}

func (bs *BitStream) ReadUint16() ([]byte, uint16) {

	bytes, ok := bs.ReadBits(2 * 8, true)
	if !ok {
		return []byte{0, 0, 0}, 0
	}
	return bytes, binary.LittleEndian.Uint16(bytes)
}

func (bs *BitStream) ReadUint32() ([]byte, uint32) {

	bytes, ok := bs.ReadBits(4 * 8, true)
	if !ok {
		return []byte{0, 0, 0}, 0
	}
	return bytes, binary.LittleEndian.Uint32(bytes)
}

func (bs *BitStream) WriteByte(b byte) {

	bs.Write([]byte{b}, 8, true)
}