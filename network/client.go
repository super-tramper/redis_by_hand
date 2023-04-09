package network

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net"
	"os"
	"strings"
)

const MaxMsg = 4096

func RunClient() {
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
		content, err := readFromServer(conn)
		if err != nil {
			fmt.Println(err)
			log.Errorf("client read error: %v", err)
		}
		fmt.Println(string(content))
	}
}

func query(conn net.Conn) error {
	ic := make([]byte, 0)
	buffer := bytes.NewBuffer(ic)
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	inputContent := strings.Trim(input, "\r\n")

	if strings.ToUpper(inputContent) == "Q" {
		os.Exit(0)
	}

	length := uint32(4 + len(inputContent))
	if length > MaxMsg {
		return fmt.Errorf("wrong frame length %d", length)
	}
	if err := binary.Write(buffer, binary.BigEndian, length); err != nil {
		return err
	}
	buffer.Write([]byte(inputContent))

	_, err := conn.Write(buffer.Bytes())
	if err != nil {
		log.Errorf("client send message failed: %v", err)
		return err
	}
	return nil
}

func readFromServer(conn net.Conn) ([]byte, error) {
	buf := make([]byte, MaxMsg+4)
	if _, err := conn.Read(buf); err != nil {
		return nil, err
	}

	buffer := bytes.NewBuffer(buf)

	var length uint32
	if err := binary.Read(buffer, binary.BigEndian, &length); err != nil {
		return nil, err
	}

	if length > 4+MaxMsg {
		return nil, fmt.Errorf("received data exceeded limitation: %d", length)
	}

	return buf[4:length], nil
}
