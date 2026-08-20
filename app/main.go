package main

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
)

// StatusCode defines a custom type for status code enum.
type HttpStatus int

const (
	OK          HttpStatus = 200
	BAD_REQUEST HttpStatus = 400
	NOT_FOUND   HttpStatus = 404
)

type HttpResponse struct {
	Version string
	Status  HttpStatus
	Headers map[string]string
	Body    string
}

type HttpRequest struct {
	Method  string
	Path    string
	Version string
	Headers map[string]string
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
		conn.Write(OKResponse())
	} else if strings.Contains(req.Path, "/echo") {
		conn.Write(OKResponse(req.GetEchoContent()))
	} else if strings.Contains(req.Path, "/user-agent") {
		headerValue := req.Headers["user-agent"]
		conn.Write(OKResponse(headerValue))
	} else {
		conn.Write(NotFoundResponse())
	}
}

/** Request helper functions  */
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

	headers := make(map[string]string)
	for i := 1; i < len(lines); i++ {
		line := strings.Split(lines[i], ": ")
		if len(line) == 2 {
			headers[strings.ToLower(line[0])] = line[1]
		}
	}

	req := &HttpRequest{
		Method:  requestLineParts[0],
		Path:    requestLineParts[1],
		Version: requestLineParts[2],
		Headers: headers,
	}

	return req, nil
}

func (r *HttpRequest) GetEchoContent() string {
	echoPath := "/echo/"
	idx := strings.Index(r.Path, echoPath)
	if idx == -1 {
		return ""
	}

	return r.Path[(idx + len(echoPath)):]
}

/** Http Status helper functions  */
func (s HttpStatus) ReasonPhrase() string {
	switch s {
	case OK:
		return "OK"
	case BAD_REQUEST:
		return "Bad Request"
	case NOT_FOUND:
		return "Not Found"
	default:
		return "Unknown"
	}
}

/** Response helper functions  */
func OKResponse(body ...string) []byte {
	defaultBody := ""
	if len(body) > 0 {
		defaultBody = body[0]
	}

	response := HttpResponse{
		Version: "HTTP/1.1",
		Status:  OK,
		Body:    defaultBody,
	}

	return response.ToJSON()
}

func NotFoundResponse() []byte {
	response := HttpResponse{
		Version: "HTTP/1.1",
		Status:  NOT_FOUND,
	}

	return response.ToJSON()
}

func (r *HttpResponse) ToJSON() []byte {
	statusLine := fmt.Sprintf("%s %d %s\r\n", r.Version, r.Status, r.Status.ReasonPhrase())

	headers := ""
	if r.Body != "" {
		headers += "Content-Type: text/plain\r\n"
		headers += fmt.Sprintf("Content-Length: %d\r\n", len(r.Body))
	}

	return []byte(statusLine + headers + "\r\n" + r.Body)
}
