package constants

import (
	"bytes"
	"encoding/binary"
	log "github.com/sirupsen/logrus"
)

type SerType int32

const (
	SerNil SerType = iota
	SerErr
	SerStr
	SerInt
	SerArr
)

func SerTypeBytes(n SerType) []byte {
	b := make([]byte, 0)
	buffer := bytes.NewBuffer(b)
	if err := binary.Write(buffer, binary.BigEndian, n); err != nil {
		log.Errorf("write serialization type error: %v", err)
	}
	return buffer.Bytes()
}
