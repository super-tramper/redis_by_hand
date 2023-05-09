package serialization

import (
	"bytes"
	"encoding/binary"
	log "github.com/sirupsen/logrus"
	"redis_by_hand/constants"
	"redis_by_hand/tools"
	"unsafe"
)

func SerializeNil(out *[]byte) {
	*out = append(*out, constants.SerTypeBytes(constants.SerNil)...)
}

func SerializeStr(out *[]byte, val *string) {
	*out = append(*out, constants.SerTypeBytes(constants.SerStr)...)
	*out = append(*out, tools.UInt32Bytes(uint32(len(*val)))...)
	*out = append(*out, []byte(*val)...)
}

func SerializeInt(out *[]byte, val int64) {
	*out = append(*out, constants.SerTypeBytes(constants.SerInt)...)
	*out = append(*out, tools.Int64Bytes(val)...)
}

func SerializeErr(out *[]byte, code constants.ErrType, msg *string) {
	*out = append(*out, constants.SerTypeBytes(constants.SerErr)...)
	*out = append(*out, tools.Int32Bytes(int32(code))...)
	*out = append(*out, []byte(*msg)...)
}

func SerializeArr(out *[]byte, n uint32) {
	*out = append(*out, constants.SerTypeBytes(constants.SerArr)...)
	*out = append(*out, tools.IntBytes(n)...)
}

func SerializeDbl(out *[]byte, val float64) {
	*out = append(*out, constants.SerTypeBytes(constants.SerDbl)...)
	*out = append(*out, tools.FloatBytes(val)...)
}

func SerializeUpdateArr(out *[]byte, n uint32) {
	if DeserializeSerType(out) != constants.SerArr {
		log.Errorf("serialize update expected arr.")
		return
	}
	var b []byte
	buffer := bytes.NewBuffer(b)
	err := binary.Write(buffer, binary.BigEndian, n)
	if err != nil {
		log.Errorf("serialize update write int error.")
		return
	}
	nBytes := buffer.Bytes()
	for i := 0; i < int(unsafe.Sizeof(uint32(0))); i++ {
		(*out)[int(unsafe.Sizeof(constants.SerArr))+i] = nBytes[i]
	}
}
