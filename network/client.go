package network

import (
	"bufio"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net"
	"os"
	"strings"
)

func RunClient() {
	conn, err := net.Dial("tcp", "127.0.0.1:1234")
	if err != nil {
		log.Errorf("create client connction error: %v", err)
		return
	}
	defer conn.Close()

	for {
		if n, err := sendToServer(conn); err != nil || n == 0 {
			log.Errorf("send to server error: %v", err)
			continue
		}
		readFromServer(conn)
	}
}

func sendToServer(conn net.Conn) (int, error) {
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	inputContent := strings.Trim(input, "\r\n")
	fmt.Printf("input content: %v\n", inputContent)

	if strings.ToUpper(inputContent) == "Q" {
		os.Exit(0)
	}

	n, err := conn.Write([]byte(inputContent))
	if err != nil {
		log.Errorf("client send message failed: %v", err)
		return 0, err
	}
	return n, nil
}

func readFromServer(conn net.Conn) {
	fmt.Println("in read")
	buf := [512]byte{}
	n, err := conn.Read(buf[:])
	if err != nil {
		log.Errorf("cilent read message failed: %v", err)
	}
	fmt.Println(string(buf[:n]))
}
