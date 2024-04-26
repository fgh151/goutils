package main

import (
	"github.com/runetid/go-sdk/models"
	"log"
)

func main() {
	req := models.InternalRequest{
		Host:   "localhost:555",
		Method: "get-user",
		Body:   "test",
	}

	resp, err := models.SockFetch[string](&req)

	if err != nil {
		log.Fatalln("Cant receive")
	}

	log.Println("Receive " + resp)

}
