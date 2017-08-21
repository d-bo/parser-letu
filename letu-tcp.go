package main

// Go letu parser service
// TODO: single mongo connection, close on SIGTERM, SIGINT

import (
	"os"
    "net"
	"fmt"
    "time"
	"syscall"
	"os/signal"
	"goldapple"
	"gopkg.in/mgo.v2"
)

const (
	C_HOST = "0.0.0.0"
	C_PORT = "8800"
	C_TYPE = "tcp"
)

const LetuBrandCollection = "letu_brands"

var LetuDB string = os.Getenv("LETU_MONGO_DB")

// A time prefix before collection name
func makeTimePrefix(coll string) string {
    t := time.Now()
    ti := t.Format("02-01-2006")
    if coll == "" {
        return ti
    }
    fin := ti + "_" + coll
    return fin
}

func main() {
	glob_session, glob_err := mgo.Dial("mongodb://localhost:27017/")
	if glob_err != nil {
		fmt.Println("Mongo connection error: ", glob_err)
	}
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	go func() {
		s := <-sigc
		switch s {
			case syscall.SIGHUP:
				fmt.Println("LETU: SIGHUP")
				glob_session.Close()
				os.Exit(0)
			case syscall.SIGINT:
				fmt.Println("LETU: SIGINT")
				glob_session.Close()
				os.Exit(0)
			case syscall.SIGTERM:
				fmt.Println("LETU: SIGTERM")
				glob_session.Close()
				os.Exit(0)
			case syscall.SIGQUIT:
				fmt.Println("LETU: SIGQUIT")
				glob_session.Close()
				os.Exit(0)
		}
	}()
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
		go handleRequest(conn, glob_session)
	}
}

// goroutine handler
func handleRequest(conn net.Conn, session *mgo.Session) {
	// todo: buffer overflow ??
	buf := make([]byte, 128)
	len, err := conn.Read(buf)
	if err != nil {
		fmt.Println("Error reading: ", err.Error())
	}
	str := string(buf[:len])

	coll := makeTimePrefix(LetuBrandCollection)
	if LetuDB == "" {
		LetuDB = "parser"
	}
	c := session.DB(LetuDB).C(coll)
	session.SetMode(mgo.Monotonic, true)
	num, err := c.Count()

	if err != nil {
		fmt.Println(err)
	}

	if num > 1 {
		fmt.Println("Allready started")
		os.Exit(0)
	}

	switch str {
		case "start":
			fmt.Println("Start LETU parser ...\n")
			goldapple.Step1(session)
			goldapple.Step2(session)
			goldapple.Step3(session)
		case "step1":
			fmt.Println("Call step1.go ...\n")
			goldapple.Step1(session)
		case "step2":
			fmt.Println("Call step2.go ...\n")
			goldapple.Step2(session)
		case "step3":
			fmt.Println("Call step3.go ...\n")
			goldapple.Step3(session)
		default:
			fmt.Println("Received msg: ", str)
	}

	strbyte := []byte(str+"\n")
	conn.Write(strbyte)

	conn.Close()
}
