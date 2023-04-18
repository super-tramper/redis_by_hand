package serialization

import (
	"bytes"
	"fmt"
	"redis_by_hand/constants"
	"redis_by_hand/tools"
)

func OnResponse(b *[]byte) {
	buffer := bytes.NewBuffer(*b)
	serType := tools.DeserializeSerType(buffer)
	switch serType {
	case constants.SerNil:
		fmt.Println("(nil)")

	case constants.SerErr:
		code := tools.DeserializeErrCode(buffer)
		msg := string(buffer.Bytes())
		fmt.Printf("(error) %d %s\n", code, msg)

	case constants.SerStr:
		msg := string(buffer.Bytes())
		fmt.Println("(str) ", msg)

	case constants.SerInt:
		val := tools.DeSerializeInt64(buffer)
		fmt.Println("(int) ", val)

	case constants.SerArr:
		c := tools.DeSerializeUint32(buffer)
		fmt.Println("(arr) len =", c)
		for i := uint32(0); i < c; i++ {
			tb := buffer.Bytes()
			OnResponse(&tb)
		}
		fmt.Println("(arr) end")

	default:
		fmt.Println("bad response")
	}
}
