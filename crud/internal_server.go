package crud

import (
	"fmt"
	"github.com/runetid/go-sdk/models"
	log2 "log"
	"net"
)

func (a Application) runInternalServer() {
	log2.Println("Listening and serving HTTP on :555")
	l, err := net.Listen("tcp4", ":555")
	defer l.Close()

	if err != nil {
		panic(err)
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}
		go handleConnection(conn)
	}
}

func testHandler() {
	log2.Println("Test handler")
}

var internalHandlers = map[string]func(){
	"test": testHandler,
}

func handleConnection(conn net.Conn) {
	// Close the connection when we're done
	defer conn.Close()

	// Read incoming data
	buf := make([]byte, 1024)
	_, err := conn.Read(buf)
	if err != nil {
		fmt.Println(err)
		return
	}

	command, err := models.DecodeRequest[string](buf)

	log2.Println("Received: " + command)

	_, ok := internalHandlers[command]
	// If the key exists
	if ok {
		// Do something
		internalHandlers[command]()
	}

	msg, err := models.DecodeBytes[string](buf)

	if err != nil {
		log2.Println("Cant decode " + err.Error())
	}

	// Print the incoming data
	fmt.Println("Received: " + msg.Body.(string))

	resp := models.InternalResponse{Body: "Test response"}
	by, err := models.EncodeBytes(&resp)

	by = append(by, []byte("\n")...)

	_, err = conn.Write(by)

	if err != nil {
		log2.Println("Cant send")
		log2.Println(err)
	}
}
