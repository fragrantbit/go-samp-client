package main

import (
	"bytes"
)

func Encrypt(buf []byte, length int, port uint32, unk int) []byte {
	var bChecksum byte
	data := make([]byte, 4)
	for i := 0; i < length; i++ {
		bData := buf[i]
		bChecksum ^= bData & byte(0xAA)
	}

	encrBuffer := [1]byte{bChecksum}

	var bufNocrc bytes.Buffer

	bufNocrc.Write([]byte(encrBuffer[:1]))
	bufNocrc.Write([]byte(buf[:]))

	for i := 1; i < length; i++ {
		bufNocrc.Bytes()[i] = sampEncrTable[bufNocrc.Bytes()[i]]
		if unk == 1 {
			bufNocrc.Bytes()[i] ^= byte(port ^ 0xCC)
		}
		unk ^= 1
	}

	copy(data, bufNocrc.Bytes())

	if length >= 5 {
		tmp := bufNocrc.Bytes()[len(data):length]
		for _, v := range tmp {
			data = append(data, v)
		}
	}
	return data
}
