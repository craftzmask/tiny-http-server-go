package main

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
)

type HttpRequest struct {
	Method  string
	Path    string
	Version string
}

// Ensures gofmt doesn't remove the "net" and "os" imports above (feel free to remove this!)
var _ = net.Listen
var _ = os.Exit

func main() {
	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	conn, err := l.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}

	req, err := ParseRequest(conn)
	if err != nil {
		fmt.Println("[ParseRequest]: ", err.Error())
		os.Exit(1)
	}

	if req.Path == "/" {
		conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
	} else {
		conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
	}
}

func ParseRequest(r io.Reader) (*HttpRequest, error) {
	buffer := make([]byte, 1024)

	n, err := r.Read(buffer)
	if err != nil {
		if err == io.EOF {
			return nil, errors.New("Client disconnected")
		} else {
			return nil, fmt.Errorf("Read error: %w", err)
		}
	}

	requestString := string(buffer[:n])
	lines := strings.Split(requestString, "\r\n")
	if len(lines) == 0 {
		return nil, errors.New("Cannot parse incoming request")
	}

	requestLineParts := strings.Split(lines[0], " ")
	if len(requestLineParts) != 3 {
		return nil, fmt.Errorf("Cannot parse the request line: %v", requestLineParts)
	}

	req := &HttpRequest{
		Method:  requestLineParts[0],
		Path:    requestLineParts[1],
		Version: requestLineParts[2],
	}

	return req, nil
}
