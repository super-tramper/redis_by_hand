package serialization

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"redis_by_hand/constants"
	"redis_by_hand/tools"
	"unsafe"
)

func OnResponse(b *[]byte) int32 {
	buffer := bytes.NewBuffer(*b)
	serType := tools.DeserializeSerType(buffer)
	t_sz := int32(unsafe.Sizeof(serType))
	switch serType {
	case constants.SerNil:
		fmt.Println("(nil)")
		return t_sz

	case constants.SerErr:
		code := tools.DeserializeErrCode(buffer)
		msg := string(buffer.Bytes())
		fmt.Printf("(error) %d %s\n", code, msg)
		return t_sz + int32(unsafe.Sizeof(code)) + int32(len(msg))

	case constants.SerStr:
		length := tools.DeSerializeUint32(buffer)
		msg := string(buffer.Next(int(length)))
		fmt.Println("(str) ", msg)
		return t_sz + int32(unsafe.Sizeof(length)) + int32(length)

	case constants.SerInt:
		val := tools.DeSerializeInt64(buffer)
		fmt.Println("(int) ", val)
		return t_sz + int32(unsafe.Sizeof(val))

	case constants.SerDbl:
		val := tools.DeSerializeDbl(buffer)
		fmt.Println("(dbl) ", val)
		return t_sz + int32(unsafe.Sizeof(val))

	case constants.SerArr:
		c := tools.DeSerializeUint32(buffer)
		fmt.Println("(arr) len =", c)
		arrBytes := int32(unsafe.Sizeof(c))
		bs := buffer.Bytes()
		for i := uint32(0); i < c; i++ {
			rc := OnResponse(&bs)
			if rc < 0 {
				return rc
			}
			arrBytes += rc
			bs = bs[rc:]
		}
		fmt.Println("(arr) end")
		return t_sz + arrBytes

	default:
		fmt.Println("bad response")
		return -1
	}
}

func DeserializeSerType(b *[]byte) constants.SerType {
	buffer := bytes.NewBuffer((*b)[:unsafe.Sizeof(constants.SerArr)])
	var t constants.SerType
	if err := binary.Read(buffer, binary.BigEndian, &t); err != nil {
		return constants.SerType(-1)
	}
	return t
}
