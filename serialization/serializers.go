package serialization

import (
	"redis_by_hand/constants"
	"redis_by_hand/tools"
)

func SerializeNil(out *[]byte) {
	*out = append(*out, constants.SerTypeBytes(constants.SerNil)...)
}

func SerializeStr(out *[]byte, val *string) {
	*out = append(*out, constants.SerTypeBytes(constants.SerStr)...)
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
