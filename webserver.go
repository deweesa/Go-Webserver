package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

const (
	SERVER_HOST = "localhost"
	SERVER_PORT = "80"
	SERVER_TYPE = "tcp"
	// BASIC_RESPONSE    = "HTTP/1.1 200 Ok\r\n\r\nRequested path: %s \r\n"
	BASIC_RESPONSE    = "HTTP/1.1 200 Ok\r\n\r\n%s"
	NOT_FOUND         = "HTTP/1.1 400 Not Found\n"
	SERVING_DIRECTORY = "./www"
)

func main() {
	fmt.Println("Hello world")
	args := os.Args[1:]
	fmt.Println(args)

	server, err := net.Listen(SERVER_TYPE, SERVER_HOST+":"+SERVER_PORT)
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}
	defer server.Close()
	fmt.Println("Listening on " + SERVER_HOST + ":" + SERVER_PORT)
	fmt.Println("Waiting for client...")
	for {
		connection, err := server.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			os.Exit(1)
		}

		fmt.Println("client connected")
		go processClient(connection)
	}
}

// TODO: Refactor this method it's getting sloppy
//
//	TODO: A bit better with the reorg, now should extract methods as we can
//	NOTE: How much can we defer? I don't like having several conn.close but it's not the worse
func processClient(connection net.Conn) {
	buffer := make([]byte, 1024)

	_, err := connection.Read(buffer)
	if err != nil {
		fmt.Println("Error reading:", err.Error())
	}

	filePath := extractPath(string(buffer))
	stat, err := os.Stat(filePath)
	if err != nil {
		fmt.Println("Error finding file:", err.Error())
		connection.Write([]byte(NOT_FOUND))
		connection.Close()
		return
	}

	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("Error opening file:", err.Error())
		connection.Write([]byte(NOT_FOUND))
		connection.Close()
		return
	}

	// NOTE: I wonder if we can chunk this for larger files. Worth a new branch for messing around
	fileContents := make([]byte, stat.Size())
	_, err = bufio.NewReader(file).Read(fileContents)
	if err != nil {
		fmt.Println("Error reading file:", err.Error())
		connection.Write([]byte(NOT_FOUND))
		connection.Close()
		return
	}

	response := fmt.Sprintf(BASIC_RESPONSE, string(fileContents))
	_, err = connection.Write([]byte(response))
	if err != nil {
		fmt.Println("Error writing to socket:", err.Error())
	}

	connection.Close()
}

func extractPath(info string) string {
	tokens := strings.Split(info, " ")
	path := tokens[1]

	if path == "/" {
		path = "/index.html"
	}

	return SERVING_DIRECTORY + path
}
