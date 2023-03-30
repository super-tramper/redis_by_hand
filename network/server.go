package network

import (
	"bufio"
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
	reader := bufio.NewReader(conn)
	var buf [128]byte
	n, err := reader.Read(buf[:])
	if err != nil {
		log.Errorf("read from client failed: %v", err)
		return "Error", err
	}
	recvStr := string(buf[:n])
	return recvStr, nil
}

func sendToClient(msg string, conn net.Conn) error {
	if _, err := conn.Write([]byte(msg)); err != nil {
		return err
	}
	return nil
}
