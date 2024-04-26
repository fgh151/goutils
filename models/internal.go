package models

import (
	"bufio"
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"net"
)

type InternalRequest struct {
	Host   string
	Method string
	Body   interface{}
}

type InternalResponse struct {
	Body  interface{}
	error error
}

func EncodeBytes[T *InternalRequest | *InternalResponse](req T) ([]byte, error) {
	var buf bytes.Buffer        // Stand-in for a network connection
	enc := gob.NewEncoder(&buf) // Will write to network.

	err := enc.Encode(&req)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), err
}

func DecodeBytes[T any](data []byte) (*InternalResponse, error) {
	buf := bytes.NewBuffer(data) // Stand-in for a network connection
	dec := gob.NewDecoder(buf)   // Will read from network.

	var resp InternalResponse

	err := dec.Decode(&resp)

	resp.Body = resp.Body.(T)

	return &resp, err
}

func DecodeRequest[T any](buf []byte) (T, error) {
	msg, err := DecodeBytes[string](buf)
	return msg.Body.(T), err
}

func SockFetch[T any](req *InternalRequest) (T, error) {
	conn, err := net.Dial("tcp4", req.Host)
	if err != nil {
		log.Fatalln("Cant dial")
	}

	data, err := EncodeBytes(req)

	if err != nil {
		log.Fatalln("Cant encode request")
	}

	// Send some data to the server
	_, err = conn.Write(data)
	if err != nil {
		fmt.Println(err)
		log.Fatalln("Cant send request")
	}

	status, err := bufio.NewReader(conn).ReadBytes('\n')
	//
	if err != nil {
		log.Fatalln("Cant receive ", err)
	}

	resp, err := DecodeBytes[T](status)

	if err != nil {
		log.Fatalln("Cant decode response " + err.Error())
	}

	return resp.Body.(T), nil
}
