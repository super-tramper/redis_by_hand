package tools

import (
	"bytes"
	"encoding/binary"
	log "github.com/sirupsen/logrus"
	"redis_by_hand/constants"
)

func StrHash(data []byte, length uint64) uint64 {
	h := uint64(0x811C9DC5)
	for i := uint64(0); i < length; i++ {
		h = (h + uint64(data[i])) * 0x01000193
	}
	return h
}

func Int64Bytes(n int64) []byte {
	b := make([]byte, 0)
	buffer := bytes.NewBuffer(b)
	if err := binary.Write(buffer, binary.BigEndian, n); err != nil {
		log.Errorf("write serialization type error: %v", err)
	}
	return buffer.Bytes()
}

func Int32Bytes(n int32) []byte {
	b := make([]byte, 0)
	buffer := bytes.NewBuffer(b)
	if err := binary.Write(buffer, binary.BigEndian, n); err != nil {
		log.Errorf("write serialization type error: %v", err)
	}
	return buffer.Bytes()
}

func IntBytes(n interface{}) []byte {
	b := make([]byte, 0)
	buffer := bytes.NewBuffer(b)
	if err := binary.Write(buffer, binary.BigEndian, n); err != nil {
		log.Errorf("write serialization type error: %v", err)
	}
	return buffer.Bytes()
}

func FloatBytes(val interface{}) []byte {
	b := make([]byte, 0)
	buffer := bytes.NewBuffer(b)
	if err := binary.Write(buffer, binary.BigEndian, val); err != nil {
		log.Errorf("write serialization type error: %v", err)
	}
	return buffer.Bytes()
}

func DeserializeSerType(b *bytes.Buffer) (t constants.SerType) {
	err := binary.Read(b, binary.BigEndian, &t)
	if err != nil {
		t = -1
	}
	return
}

func DeserializeErrCode(b *bytes.Buffer) (t constants.ErrType) {
	err := binary.Read(b, binary.BigEndian, &t)
	if err != nil {
		t = -1
	}
	return
}

func DeSerializeInt64(b *bytes.Buffer) (t int64) {
	err := binary.Read(b, binary.BigEndian, &t)
	if err != nil {
		t = -1
	}
	return
}

func DeSerializeUint32(b *bytes.Buffer) (t uint32) {
	err := binary.Read(b, binary.BigEndian, &t)
	if err != nil {
		t = 0
	}
	return
}

func Max(a uint32, b uint32) uint32 {
	if a > b {
		return a
	}
	return b
}

func Min(a uint32, b uint32) uint32 {
	if a > b {
		return b
	}
	return a
}

func BToI(b bool) int {
	if b {
		return 1
	}
	return 0
}
