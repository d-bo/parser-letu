package main

// Go letu parser service
// TODO: single mongo connection, close on SIGTERM, SIGINT

import (
	"os"
    "net"
    "fmt"
	"goldapple"
)

const (
	C_HOST = "0.0.0.0"
	C_PORT = "8800"
	C_TYPE = "tcp"
)

func main() {
	l, err := net.Listen(C_TYPE, C_HOST+":"+C_PORT)
	if err != nil {
		fmt.Println("TCP server error: ", err)
		os.Exit(1)
	}
	defer l.Close()
	fmt.Println("Launch LETU tcp: "+C_HOST+":"+C_PORT+" ...")
	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			os.Exit(1)
		}
		go handleRequest(conn)
	}
}

// goroutine handler
func handleRequest(conn net.Conn) {
	// todo: buffer overflow ??
	buf := make([]byte, 128)
	len, err := conn.Read(buf)
	if err != nil {
		fmt.Println("Error reading: ", err.Error())
	}
	str := string(buf[:len])

	switch str {
		case "start":
			fmt.Println("Start LETU parser ...\n")
			goldapple.Step1()
			goldapple.Step2()
			goldapple.Step3()
		case "step1":
			fmt.Println("Call step1.go ...\n")
			goldapple.Step1()
		case "step2":
			fmt.Println("Call step2.go ...\n")
			goldapple.Step2()
		case "step3":
			fmt.Println("Call step3.go ...\n")
			goldapple.Step3()
		default:
			fmt.Println("Received msg: ", str)
	}

	strbyte := []byte(str+"\n")
	conn.Write(strbyte)

	conn.Close()
}
