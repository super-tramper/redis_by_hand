package network

import (
	"bytes"
	"encoding/binary"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net"
)

const maxMsg = 4096

type bufArray [4 + maxMsg + 1]byte

func RunServer() {
	addr := ":1234"
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Errorf("runserver error: %v", err)
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Infof("accept error: %v", err)
			continue
		}
		defer conn.Close()
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	// 处理连接产生panic时从错误恢复，保障server能正常运行
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("panic: %v", err)
		}
	}()

	for {
		received, err := readFromClient(conn)
		if err != nil {
			log.Errorf("read from client error: %v", err)
			panic(err)
		}
		msg := "echo: " + received
		if err := sendToClient(msg, conn); err != nil {
			log.Errorf("send to client error: %v", err)
			panic(err)
		}
	}
}

func readFromClient(conn net.Conn) (string, error) {

	//reader := bufio.NewReader(conn)
	//var buf [128]byte
	//n, err := reader.Read(buf[:])
	//if err != nil {
	//	log.Errorf("read from client failed: %v", err)
	//	return "Error", err
	//}
	//recvStr := string(buf[:n])
	//return recvStr, nil
	for {
		if err := oneRequest(conn); err != nil {
			return "", err
		}
	}
}

func sendToClient(msg string, conn net.Conn) error {
	if _, err := conn.Write([]byte(msg)); err != nil {
		return err
	}
	return nil
}

func readFull(conn net.Conn, buf []byte, n uint32) (int32, error) {
	for total := uint32(0); n > total; {
		rn, err := conn.Read(buf[total:])
		if err != nil || rn <= 0 {
			return -1, err
		}
		if total > n {
			panic("read exceeded")
		}
		total += uint32(rn)
	}
	return 0, nil
}

func writeAll(conn net.Conn, buf []byte, n uint32) (int32, error) {
	for rv := 0; n > 0; n -= uint32(rv) {
		w, err := conn.Write(buf)
		if err != nil || w <= 0 {
			return -1, err
		}
		if uint32(rv) > n {
			panic("read exceeded")
		}
		rv += w
	}
	return 0, nil
}

func oneRequest(conn net.Conn) error {
	var b []byte
	bytesBuffer := bytes.NewBuffer(b)

	if _, err := readFull(conn, bytesBuffer.Bytes(), 4); err != nil {
		return err
	}
	if _, err := bytesBuffer.ReadFrom(conn); err != nil {
		return err
	}

	var length uint32 = 0
	if err := binary.Read(bytesBuffer, binary.BigEndian, &length); err != nil {
		return err
	}

	if length > maxMsg {
		return fmt.Errorf("too long")
	}

	if _, err := readFull(conn, bytesBuffer.Bytes(), length); err != nil {
		return err
	}

	//bytesBuffer.Write([]byte{'\0'})
	fmt.Println("client says: ", bytesBuffer.Bytes()[4:])

	reply := []byte{'w', 'o', 'r', 'l', 'd'}
	var w []byte
	wBytesBuffer := bytes.NewBuffer(w)
	length = uint32(len(reply))
	if err := binary.Write(wBytesBuffer, binary.BigEndian, &length); err != nil {
		return err
	}
	if err := binary.Write(wBytesBuffer, binary.BigEndian, reply); err != nil {
		return err
	}

	_, err := writeAll(conn, wBytesBuffer.Bytes()[:], 4+length)
	return err
}
