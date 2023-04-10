package packet

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// Packet协议
/*
# req packet
nstr 字符串数量 32位整数
len 字符串长度 32位整数
str 字符串

# res
status 32位状态码
data 响应字符串
*/

type Packet interface {
	Decode([]byte) error     // []byte -> struct
	Encode() ([]byte, error) //  struct -> []byte
}

type ReqBody struct {
	StrLen int32
	Str    string
}

type ReqPacket struct {
	StrCnt  int32
	Payload []*ReqBody
}

type ResPacket struct {
	Status int32
	Data   string
}

func (req *ReqPacket) Decode(payload []byte) error {
	buffer := bytes.NewBuffer(payload)
	// read StrCnt
	if err := binary.Read(buffer, binary.BigEndian, req.StrCnt); err != nil {
		return err
	}
	// read body
	for i := int32(0); i < req.StrCnt; i++ {
		if err := req.DecodeBody(buffer); err != nil {
			return err
		}
	}
	return nil
}

func (req *ReqPacket) Encode() ([]byte, error) {
	b := make([]byte, 0)
	buffer := bytes.NewBuffer(b)

	// write StrCnt
	if err := binary.Write(buffer, binary.BigEndian, req.StrCnt); err != nil {
		return nil, fmt.Errorf("ReqPacket Encode error: %v", err)
	}

	// write body
	if err := req.EncodeBody(buffer); err != nil {
		return nil, fmt.Errorf("ReqPacket Encode error: %v", err)
	}

	return buffer.Bytes(), nil
}

func (req *ReqPacket) DecodeBody(buffer *bytes.Buffer) error {
	var body *ReqBody
	if err := binary.Read(buffer, binary.BigEndian, body.StrLen); err != nil {
		return err
	}
	body.Str = string(buffer.Next(int(body.StrLen)))
	if len(body.Str) != int(body.StrLen) {
		return fmt.Errorf("expected length: %d, got %s", body.StrLen, body.Str)
	}
	req.Payload = append(req.Payload, body)
	return nil
}

func (req *ReqPacket) EncodeBody(buffer *bytes.Buffer) error {
	for i := int32(0); i < req.StrCnt; i++ {
		reqBody := req.Payload[i]
		// 写入长度
		if err := binary.Write(buffer, binary.BigEndian, reqBody.StrLen); err != nil {
			return err
		}

		// 写入字符串
		buffer.Write([]byte(reqBody.Str))
	}
	return nil
}

func (res *ResPacket) Decode(payload []byte) error {
	buffer := bytes.NewBuffer(payload)
	// read Status
	if err := binary.Read(buffer, binary.BigEndian, res.Status); err != nil {
		return err
	}
	// read Data
	res.Data = buffer.String()
	return nil
}

func (res *ResPacket) Encode() ([]byte, error) {
	b := make([]byte, 0)
	buffer := bytes.NewBuffer(b)
	// 写入状态码
	if err := binary.Write(buffer, binary.BigEndian, res.Status); err != nil {
		return nil, err
	}
	// 写入字符串
	buffer.Write([]byte(res.Data))

	return buffer.Bytes(), nil
}
