package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net"
	"os"
	"redis_by_hand/network/frame"
	"redis_by_hand/network/packet"
	"strings"
)

const MaxMsg = 4096
const PORT = 1234

func main() {
	addr := fmt.Sprintf("127.0.0.1:%d", PORT)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		log.Errorf("create client connction error: %v", err)
		return
	}
	defer conn.Close()

	for {
		if err := query(conn); err != nil {
			log.Errorf("send to server error: %v", err)
			continue
		}
		resp, err := readFromServer(conn)
		if err != nil {
			log.Errorf("client read error: %v", err)
			continue
		}
		fmt.Println("result: status", resp.Status, "data", resp.Data)
	}
}

func query(conn net.Conn) error {
	req, err := readFromRepl()
	if err != nil {
		return err
	}

	var reqFrame frame.Frame
	buffer := bytes.NewBuffer(reqFrame)
	length := int32(len(req) + 4)
	if err = binary.Write(buffer, binary.BigEndian, &length); err != nil {
		return err
	}
	buffer.Write(req)

	_, err = conn.Write(buffer.Bytes())
	if err != nil {
		log.Errorf("client send message failed: %v", err)
		return err
	}
	return nil
}

// 从输入中读取内容，构造reqPacket
func readFromRepl() ([]byte, error) {
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')

	cmd := strings.Split(input, " ")

	req := &packet.ReqPacket{}
	req.StrCnt = int32(len(cmd))

	for _, c := range cmd {
		content := strings.Trim(c, "\r\n")
		reqBody := &packet.ReqBody{
			StrLen: int32(len(content)),
			Str:    content,
		}
		req.Payload = append(req.Payload, reqBody)
	}

	return req.Encode()
}

func readFromServer(conn net.Conn) (*packet.ResPacket, error) {
	framePayload, err := DecodeFrame(conn)
	if err != nil {
		return nil, err
	}

	resp := &packet.ResPacket{}
	if err := resp.Decode(framePayload); err != nil {
		return nil, err
	}

	return resp, nil
}

func DecodeFrame(conn net.Conn) ([]byte, error) {
	b := make([]byte, MaxMsg+4)
	if _, err := conn.Read(b); err != nil {
		return nil, err
	}

	buffer := bytes.NewBuffer(b)
	header := int32(0)
	if err := binary.Read(buffer, binary.BigEndian, &header); err != nil {
		return nil, err
	}
	return b[4:header], nil
}
